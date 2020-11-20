package scannetwork

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/hophouse/gop/utils"
)

type hostStruct struct {
	ip       net.IP
	services []serviceStruct
}

type serviceStruct struct {
	port       int
	portString string
	protocol   string
	status     string
}

// RunScanNetwork will run network scan on all inputed IP
func RunScanNetwork(inputFile *os.File, tcpOption bool, udpOption bool, portsString string, onlyOpen bool, concurrency int) {
	workersChan := make(chan bool)
	ipChan := make(chan net.IP, concurrency)
	gatherChan := make(chan hostStruct)
	results := make([]hostStruct, 0)

	// Parse ports
	ports := parsePortsOption(portsString)
	if len(ports) < 1 {
		utils.Log.Println("No valid port found. Exiting.")
	}

	// Run workers
	for i := 0; i < concurrency; i++ {
		go scanWorker(ipChan, workersChan, gatherChan, tcpOption, udpOption, ports)
	}
	defer close(workersChan)

	// Run goroutine to gather results and add them to the result slice
	go func(gatherChan chan hostStruct, results *[]hostStruct, workersChan chan bool) {
		for item := range gatherChan {
			*results = append(*results, item)
		}
		workersChan <- true
	}(gatherChan, &results, workersChan)

	// Parse IP addresses
	scanner := bufio.NewScanner(inputFile)
	for scanner.Scan() {
		ipAddr := net.ParseIP(scanner.Text())
		if ipAddr == nil {
			/*
				// Let's try to parse it as CIDR
				//ipAddr, ipNet, err := net.ParseCIDR(scanner.Text())
				_, _, err := net.ParseCIDR(scanner.Text())
				if err != nil {
					break
				}
			*/
			break
		}

		ipChan <- ipAddr
	}
	// Close IP chan
	close(ipChan)

	// Wait for GoRoutine to finish
	for i := 0; i < concurrency; i++ {
		<-workersChan
	}

	// Wait for gather worker to finish
	close(gatherChan)

	// Wait for the gather worker to finish
	<-workersChan

	// Parse results
	for _, item := range results {
		w := tabwriter.NewWriter(os.Stdout, 4, 1, 4, ' ', 0)
		fmt.Printf("[+] %s :\n", item.ip.String())

		for _, service := range item.services {
			if onlyOpen && service.status == "Open" {
				fmt.Fprintf(w, "\t%s/%s\t%s\n", service.protocol, service.portString, service.status)
				break
			}
			fmt.Fprintf(w, "\t%s/%s\t%s\n", service.protocol, service.portString, service.status)
		}

		w.Flush()
	}
}

func scanWorker(ipChan chan net.IP, workersChan chan bool, gatherChan chan hostStruct, tcpOption bool, udpOtion bool, ports []int) {
	for ipAddr := range ipChan {
		host := hostStruct{
			ip: ipAddr,
		}

		for _, portInt := range ports {
			service := serviceStruct{}
			service.port = portInt
			service.portString = strconv.Itoa(portInt)

			address := fmt.Sprintf("%s:%s", ipAddr.String(), service.portString)

			if tcpOption {
				service.protocol = "TCP"
				service.status = "Close"

				_, err := net.Dial("tcp", address)
				if err == nil {
					service.status = "Open"
				}

				host.services = append(host.services, service)
			}

			if udpOtion {
				service.protocol = "UDP"
				service.status = "Close"

				_, err := net.Dial("udp", address)
				if err == nil {
					service.status = "Open"
				}

				host.services = append(host.services, service)
			}
		}
		gatherChan <- host
	}
	workersChan <- true
}

func parsePortsOption(portsOption string) []int {
	var result []int
	/*
		Possible forms :
		X,Y,Z
		X-Y
		X-Y,Z
	*/
	for _, i := range strings.Split(portsOption, ",") {
		item := strings.TrimSpace(i)
		if itemDash := strings.Split(item, "-"); len(itemDash) == 2 {
			i1, err := strconv.ParseInt(itemDash[0], 10, 32)
			if err != nil {
				break
			}
			i2, err := strconv.ParseInt(itemDash[1], 10, 32)
			if err != nil {
				break
			}

			if i1 > i2 {
				break
			}

			for c := i1; c <= i2; c++ {
				result = append(result, int(c))
			}

		} else {
			i3, err := strconv.ParseInt(i, 10, 32)
			if err != nil {
				break
			}
			result = append(result, int(i3))
		}
	}
	return result
}
