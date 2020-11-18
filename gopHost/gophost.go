package gophost

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"text/tabwriter"
)

type register struct {
	domain string
	ip     []string
}

// RunHostCmd : Run the Host command and do a DNS lookup for each host inputed
func RunHostCmd(reader *os.File, concurrency int) {
	// Scanner to read file
	scanner := bufio.NewScanner(reader)
	workersChan := make(chan bool)
	domainsChan := make(chan string, concurrency)
	gatherChan := make(chan register)
	results := make([]register, 0)

	for i := 0; i < concurrency; i++ {
		go worker(domainsChan, gatherChan, workersChan)
	}
	defer close(workersChan)

	// Consumme the results from the worker
	go func() {
		for elem := range gatherChan {
			results = append(results, elem)
		}
		workersChan <- true
	}()

	// Read domains, add them to the channel, then close it
	for scanner.Scan() {
		domainsChan <- scanner.Text()
	}
	close(domainsChan)

	// Wait for workers to finish
	for i := 0; i < concurrency; i++ {
		<-workersChan
	}

	// Close the channel and then wait for the gather worker to finish
	close(gatherChan)
	// Last worker is the one that consumme workers' results and adf them to the results slice
	<-workersChan

	// Display results
	w := tabwriter.NewWriter(os.Stdout, 8, 4, 2, ' ', 0)
	for _, elem := range results {
		ipList := strings.Join(elem.ip, " ")
		fmt.Fprintf(w, "%s\t%s\n", elem.domain, ipList)
	}
	w.Flush()
}

func worker(domainsChan chan string, gatherChan chan register, workersChan chan bool) {
	for domain := range domainsChan {
		ipList, _ := net.LookupHost(domain)
		gatherChan <- register{
			domain: domain,
			ip:     ipList,
		}
	}
	workersChan <- true
}
