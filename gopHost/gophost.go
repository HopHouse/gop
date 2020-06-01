package gophost

import (
    "bufio"
	"fmt"
	"net"
	"os"
)

func RunHostCmd(reader *os.File) {
    // Scanner to read file
	scanner := bufio.NewScanner(reader)
    for scanner.Scan() {
		domain := scanner.Text()
		ip_list, _ := net.LookupHost(domain)
		for _, ip := range ip_list {
			fmt.Printf("%s %s\n", domain, ip)
		}
	}
}