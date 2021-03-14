package scannetwork

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/hophouse/gop/utils"
)

type hostStruct struct {
	mu              sync.Mutex
	ip              net.IP
	services        []serviceStruct
	hasOpenServices bool
}

type serviceStruct struct {
	ip         string
	port       int
	portString string
	protocol   string
	status     string
}

// RunScanNetwork will run network scan on all inputed IP
func RunScanNetwork(inputFile *os.File, tcpOption bool, udpOption bool, portsString string, onlyOpen bool, concurrency int, output string) {
	workersChan := make(chan bool)
	inputChan := make(chan string, concurrency)
	gatherChan := make(chan serviceStruct)
	hosts := make(map[string]hostStruct, 0)

	// Init ports list
	initPortStringMap()

	// Parse ports
	ports := unique(parsePortsOption(portsString))
	if len(ports) < 1 {
		utils.Log.Println("No valid port found. Exiting.")
	}

	// Run workers
	for i := 0; i < concurrency; i++ {
		go scanWorker(inputChan, workersChan, gatherChan, tcpOption, udpOption)
	}

	// Run goroutine to gather results and add them to the result slice
	go func(gatherChan chan serviceStruct, workersChan chan bool, hosts map[string]hostStruct) {
		for service := range gatherChan {
			host, _ := hosts[service.ip]
			host.mu.Lock()
			host.services = append(host.services, service)

			if service.status == "Open" {
				host.hasOpenServices = true
			}
			host.mu.Unlock()
			hosts[service.ip] = host
		}
		workersChan <- true
	}(gatherChan, workersChan, hosts)

	// Parse IP addresses
	scanner := bufio.NewScanner(inputFile)
	for scanner.Scan() {
		if !strings.Contains(scanner.Text(), "/") {
			ipAddr := net.ParseIP(scanner.Text())
			if ipAddr == nil {
				utils.Log.Printf("[!] Input %s is not a valid IP address.\n", scanner.Text())
				fmt.Printf("[!] Input %s is not a valid IP address.\n", scanner.Text())
				break
			}
			hosts[ipAddr.String()] = hostStruct{
				ip:              ipAddr,
				services:        []serviceStruct{},
				hasOpenServices: false,
			}
			utils.Log.Printf("[+] Adding IP address to queue : %s\n", scanner.Text())
		} else {
			ipAddrs, err := ipsFromCIDR(scanner.Text())
			if err != nil {
				utils.Log.Printf("[!] Input %s is not a valid IP address range.\n", scanner.Text())
				fmt.Printf("[!] Input %s is not a valid IP address range.\n", scanner.Text())
				break
			}

			for _, ipAddr := range ipAddrs {
				hosts[ipAddr] = hostStruct{
					ip:              net.ParseIP(ipAddr),
					services:        []serviceStruct{},
					hasOpenServices: false,
				}
				utils.Log.Printf("[+] Adding IP address to queue : %s\n", ipAddr)
			}
		}
	}

	// Run the scan
	for host, _ := range hosts {
		for _, port := range ports {
			inputChan <- fmt.Sprintf("%s:%d", host, port)
		}
	}
	// Close IP chan
	close(inputChan)

	// Wait for GoRoutine to finish
	for i := 0; i < concurrency; i++ {
		<-workersChan
	}

	// Wait for gather worker to finish
	close(gatherChan)

	// Wait for the gather worker to finish
	<-workersChan

	switch output {
	case "grep":
		printGreppableResults(hosts, onlyOpen)
	case "short":
		printShortResults(hosts, onlyOpen)
	default:
		printResults(hosts, onlyOpen)
	}
	utils.Log.Println("[+] Scan is terminated")
}

func printResults(hosts map[string]hostStruct, onlyOpen bool) {
	for ip, item := range hosts {
		if item.hasOpenServices == false && onlyOpen == true {
			continue
		}

		fmt.Printf("\n[+] %s :\n", ip)
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
}

func printGreppableResults(hosts map[string]hostStruct, onlyOpen bool) {
	for ip, item := range hosts {
		if item.hasOpenServices == false && onlyOpen == true {
			continue
		}

		for _, service := range item.services {
			if strings.Compare(service.status, "Open") == 0 {
				fmt.Printf("%s,%s,%s,%s\n", ip, service.protocol, service.portString, service.status)
			} else {
				if onlyOpen == false {
					fmt.Printf("%s,%s,%s,%s\n", ip, service.protocol, service.portString, service.status)
				}
			}
		}
	}
}

func printShortResults(hosts map[string]hostStruct, onlyOpen bool) {
	for ip, item := range hosts {
		if item.hasOpenServices == false && onlyOpen == true {
			continue
		}
		for _, service := range item.services {
			if strings.Compare(service.status, "Open") == 0 {
				fmt.Printf("%s:%s\n", ip, service.portString)
			}
		}
	}
}
func scanWorker(inputChan chan string, workersChan chan bool, gatherChan chan serviceStruct, tcpOption bool, udpOtion bool) {
	for entry := range inputChan {
		service := serviceStruct{
			ip:         "",
			port:       0,
			portString: "",
			protocol:   "",
			status:     "",
		}
		service.ip = strings.Split(entry, ":")[0]
		service.port, _ = strconv.Atoi(strings.Split(entry, ":")[1])
		service.portString = strconv.Itoa(service.port)

		address := fmt.Sprintf("%s:%s", service.ip, service.portString)

		if tcpOption {
			service.protocol = "TCP"
			service.status = "Close"

			conn, err := net.DialTimeout("tcp", address, 500*time.Millisecond)
			if err == nil {
				service.status = "Open"
				conn.Close()
			}
			gatherChan <- service
		}

		if udpOtion {
			service.protocol = "UDP"
			service.status = "Close"

			conn, err := net.Dial("udp", address)
			if err == nil {
				service.status = "Open"
				conn.Close()
			}
			gatherChan <- service
		}
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
		// Check if a string that represent a set of ports is passed
		portList, in := portStringMap[item]
		if in == true {
			item = strings.TrimSpace(portList)

			// Recursively parse ports
			result = append(result, parsePortsOption(item)...)

			continue
		}

		// No string is passed, then parse the ports
		if itemDash := strings.Split(item, "-"); len(itemDash) == 2 {
			i1, err := strconv.ParseInt(itemDash[0], 10, 32)
			if err != nil {
				continue
			}
			i2, err := strconv.ParseInt(itemDash[1], 10, 32)
			if err != nil {
				continue
			}

			if i1 > i2 {
				continue
			}

			for c := i1; c <= i2; c++ {
				result = append(result, int(c))
			}

		} else {
			i3, err := strconv.ParseInt(item, 10, 32)
			if err != nil {
				continue
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

func unique(intSlice []int) []int {
	keys := make(map[int]bool)
	list := []int{}
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
