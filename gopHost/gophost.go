package gophost

import (
	"bufio"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	whois "github.com/hophouse/golang-whois"
	"github.com/hophouse/gop/utils/logger"
	"github.com/miekg/dns"
)

type register struct {
	domain            string
	middleInformation []string
	ip                []string
}

var DnsServers = []string{
	"1.1.1.1:53", "1.0.0.1:53", // CloudFlare
	"208.67.222.222:53", "208.67.220.220:53", // OpenDNS
	"9.9.9.10:53", "149.112.112.10:53", // Quad9
	"84.200.69.80:53", "84.200.70.40:53", // DNS.Watch
}

var AvailabilityCheck bool

// RunHostCmd : Run the Host command and do a DNS lookup for each host inputed
func RunHostCmd(reader *os.File, concurrency int, petitPoucet bool, availabilityCheckOption bool) {
	AvailabilityCheck = availabilityCheckOption

	// Scanner to read file
	scanner := bufio.NewScanner(reader)
	workersChan := make(chan bool)
	domainsChan := make(chan string, concurrency)
	gatherChan := make(chan register)
	results := make([]register, 0)
	//tldAvailability := make()

	for i := 0; i < concurrency; i++ {
		if !petitPoucet {
			go workerLookup(domainsChan, gatherChan, workersChan)
		} else {
			go workerLookupPetitPoucet(domainsChan, gatherChan, workersChan)
		}
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
		domain := scanner.Text()
		domainCleaned := strings.TrimSpace(domain)
		domainsChan <- domainCleaned
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

	if !petitPoucet {
		// Display results
		w := tabwriter.NewWriter(os.Stdout, 8, 4, 2, ' ', 0)
		for _, elem := range results {
			// formt IP results
			ipList := strings.Join(elem.ip, " ")
			domain := elem.domain

			if AvailabilityCheck {
				// format domain with availability data
				available, err := checkAvailability(domain)
				if err != nil {
					logger.Printf("Error trying to whois the domaine. %s\n", err)
				}
				if available == true {
					domain = fmt.Sprintf("[FREE] %s", domain)
				}
			}

			logger.Fprintf(w, "%s\t%s\n", domain, ipList)
		}
		w.Flush()
	} else {
		for _, elem := range results {
			resultsSlice := make([]string, 0)

			domain := elem.domain

			if AvailabilityCheck {
				// format domain with availability data
				available, err := checkAvailability(domain)
				if err != nil {
					logger.Printf("Error trying to whois the domaine. %s\n", err)
				}
				if available == true {
					domain = fmt.Sprintf("[FREE] %s", domain)
				}
			}
			resultsSlice = append(resultsSlice, domain)

			if len(elem.middleInformation) > 1 {
				for _, domain := range elem.middleInformation {
					if AvailabilityCheck {
						available, err := checkAvailability(domain)
						if err != nil {
							logger.Printf("Error trying to whois the domaine. %s\n", err)
						}
						if available == true {
							domain = fmt.Sprintf("[FREE] %s", domain)
						}
					}
					resultsSlice = append(resultsSlice, domain)
				}
			}

			if len(elem.ip) > 0 {
				ipList := strings.Join(elem.ip, " ")
				resultsSlice = append(resultsSlice, ipList)
			}

			results := strings.Join(resultsSlice, " > ")

			logger.Printf("%s\n", results)
		}
	}

}

func workerLookup(domainsChan chan string, gatherChan chan register, workersChan chan bool) {
	for domain := range domainsChan {
		ipList, _ := net.LookupHost(domain)
		ipListUnique := uniqueNonEmptyElementsOf(ipList)
		gatherChan <- register{
			domain:            domain,
			middleInformation: []string{},
			ip:                ipListUnique,
		}
	}
	workersChan <- true
}

func uniqueNonEmptyElementsOf(s []string) []string {
	unique := make(map[string]bool, len(s))
	result := make([]string, len(unique))
	for _, elem := range s {
		if len(elem) != 0 {
			if !unique[elem] {
				result = append(result, elem)
				unique[elem] = true
			}
		}
	}
	return result
}

func workerLookupPetitPoucet(domainsChan chan string, gatherChan chan register, workersChan chan bool) {

	rand.Seed(time.Now().UnixNano())

	for domain := range domainsChan {
		dnsServer := DnsServers[rand.Intn(len(DnsServers))]

		cname := domain

		register := register{
			domain:            domain,
			middleInformation: []string{},
			ip:                []string{},
		}

		for {
			cnameList, err := lookupCNAME(cname, dnsServer)
			if err == nil && len(cnameList) > 0 {
				cname = cnameList[0]
				register.middleInformation = append(register.middleInformation, cname)
				continue
			}

			ipList, err := lookupA(cname, dnsServer)
			if err == nil && len(ipList) > 0 {
				register.middleInformation = append(register.middleInformation, cname)
				register.ip = ipList
			}
			break
		}

		gatherChan <- register
	}

	workersChan <- true
}

func lookupA(domain string, dnsServer string) ([]string, error) {
	var msg dns.Msg
	ipList := make([]string, 0)

	fqdn := dns.Fqdn(domain)
	msg.SetQuestion(fqdn, dns.TypeA)

	in, err := dns.Exchange(&msg, dnsServer)
	if err != nil {
		return ipList, errors.New("Server could not be contacted")
	}

	if len(in.Answer) < 1 {
		return ipList, errors.New("No A records")
	}

	for _, answer := range in.Answer {
		if a, ok := answer.(*dns.A); ok {
			ipList = append(ipList, a.A.String())
		}
	}

	return ipList, nil
}

func lookupCNAME(domain string, dnsServer string) ([]string, error) {
	var msg dns.Msg
	var cnameList []string

	fqdn := dns.Fqdn(domain)
	msg.SetQuestion(fqdn, dns.TypeA)

	in, err := dns.Exchange(&msg, dnsServer)
	if err != nil {
		return cnameList, errors.New("Server could not be contacted")
	}

	if len(in.Answer) < 1 {
		return cnameList, errors.New("No CNAME records")
	}

	for _, answer := range in.Answer {
		if c, ok := answer.(*dns.CNAME); ok {
			cnameList = append(cnameList, c.Target)
		}
	}

	return cnameList, nil
}

func checkAvailability(domain string) (bool, error) {
	// if doamin ends with a dot, remove it. Whois do not handle it.
	if domain[len(domain)-1] == '.' {
		domain = domain[0 : len(domain)-1]
	}

	// take the domain
	topDomain := strings.Split(domain, ".")
	if len(topDomain) < 1 {
		errorString := fmt.Sprintf("Top domain are not accepted : %v ", domain)
		return false, errors.New(errorString)
	}
	tldDomain := topDomain[len(topDomain)-2] + "." + topDomain[len(topDomain)-1]

	// Avoid having the message %% Too many requests...
	var result string
	var err error
	cpt := 1
	for {
		result, err = whois.GetWhoisTimeout(tldDomain, time.Second*5)
		if err != nil {
			errorString := fmt.Sprintf("Error in whois lookup : %v ", err)
			return false, errors.New(errorString)
		}

		if strings.Contains(result, "Too many requests...") {
			time.Sleep(time.Second * ((time.Duration)(cpt)))
			cpt = cpt + 1
			continue
		} else {
			break
		}
	}

	// No Status on the domain
	resultStatus := whois.ParseDomainStatus(result)
	resultNameServer := whois.ParseNameServers(result)
	if len(resultStatus) == 0 {
		// Check if nameservers are associated
		if len(resultNameServer) == 0 {
			return true, nil
		}
	}

	for _, status := range resultStatus {
		if strings.ToLower(status) == "available" {
			return true, nil
		}
	}

	return false, nil
}
