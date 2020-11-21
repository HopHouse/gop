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
	ip              net.IP
	services        []serviceStruct
	hasOpenServices bool
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
	go func(gatherChan chan hostStruct, workersChan chan bool) {
		for item := range gatherChan {
			printResult(item, onlyOpen)
		}
		workersChan <- true
	}(gatherChan, workersChan)

	// Parse IP addresses
	scanner := bufio.NewScanner(inputFile)
	for scanner.Scan() {
		if !strings.Contains(scanner.Text(), "/") {
			ipAddr := net.ParseIP(scanner.Text())
			if ipAddr == nil {
				break
			}
			ipChan <- ipAddr
		} else {
			ipAddrs, err := ipsFromCIDR(scanner.Text())
			if err != nil {
				break
			}

			for _, ipAddr := range ipAddrs {
				ipChan <- net.ParseIP(ipAddr)
			}
		}

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
}

func printResult(item hostStruct, onlyOpen bool) {
	fmt.Printf("\n[+] %s :\n", item.ip.String())

	if item.hasOpenServices == false {
		fmt.Printf("    No open ports found\n")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 4, 1, 4, ' ', 0)
	for _, service := range item.services {
		if strings.Compare(service.status, "Open") == 0 {
			fmt.Fprintf(w, "\t%s/%s\t%s\n", service.protocol, service.portString, service.status)
		} else {
			if onlyOpen == false {
				fmt.Fprintf(w, "\t%s/%s\t%s\n", service.protocol, service.portString, service.status)
			}
		}
	}
	w.Flush()
}

func scanWorker(ipChan chan net.IP, workersChan chan bool, gatherChan chan hostStruct, tcpOption bool, udpOtion bool, ports []int) {
	for ipAddr := range ipChan {
		host := hostStruct{
			ip:              ipAddr,
			hasOpenServices: false,
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
					host.hasOpenServices = true
				}

				host.services = append(host.services, service)
			}

			if udpOtion {
				service.protocol = "UDP"
				service.status = "Close"

				_, err := net.Dial("udp", address)
				if err == nil {
					service.status = "Open"
					host.hasOpenServices = true
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

func ipsFromCIDR(cidr string) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}
	return ips, nil
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
