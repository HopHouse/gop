package gophost

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"text/tabwriter"
)

type register struct {
	domain string
	ip     []string
}

// RunHostCmd : Run the Host command and do a DNS lookup for each host inputed
func RunHostCmd(reader *os.File, concurrency int) {
	var mutex sync.RWMutex

	// Scanner to read file
	scanner := bufio.NewScanner(reader)
	workersChan := make(chan bool)
	domainsChan := make(chan string, concurrency)
	results := make([]register, 0)

	for i := 0; i < concurrency; i++ {
		go func() {
			for domain := range domainsChan {
				ipList, err := net.LookupHost(domain)
				if err != nil {
					fmt.Printf("Error for domain %s : %s", domain, err)
				}
				mutex.Lock()
				results = append(results, register{
					domain: domain,
					ip:     ipList,
				})
				mutex.Unlock()
			}
			workersChan <- true
		}()
	}
	defer close(workersChan)

	// Read domains, add them to the channel, then close it
	for scanner.Scan() {
		domainsChan <- scanner.Text()
	}
	close(domainsChan)

	// Wait for workers to finish
	for i := 0; i < concurrency; i++ {
		<-workersChan
	}

	// Consumme results
	w := tabwriter.NewWriter(os.Stdout, 8, 8, 0, '\t', 0)
	for _, elem := range results {
		ipList := strings.Join(elem.ip, " ")
		fmt.Fprintf(w, "%s\t%s\n", elem.domain, ipList)
	}
	w.Flush()
}
