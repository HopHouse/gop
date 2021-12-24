/*
Copyright © 2020 Hophouse <contact@hophouse.fr>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"os"
	"path"

	gopscannetwork "github.com/hophouse/gop/gopScanNetwork"
	"github.com/spf13/cobra"
)

var (
	tcpScanOption  bool
	udpScanOption  bool
	onlyOpenOption bool
	outputOption   string
)

// gopstaticcrawler represents the active command
var scanNetworkCmd = &cobra.Command{
	Use:   "network",
	Short: "Port scan the network. Only valid IP address must be passed as input.",
	Long:  "Port scan the network. Only valid IP address must be passed as input.",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.PersistentPreRun(cmd, args)

		var err error

		// Parse options
		reader = os.Stdin

		stat, err := reader.Stat()
		if err != nil {
			fmt.Println("Error found in stdin:", err)
			os.Exit(2)
		}

		if inputFileOption != "" {
			if (stat.Mode() & os.ModeNamedPipe) != 0 {
				fmt.Println("[!] Cannot use stdin and input-file at the same time.")
				os.Exit(2)
			}
			reader, err = os.Open(path.Join("..", inputFileOption))
			if err != nil {
				panic(err)
			}
		} else {
			// Chech if there is something in stdin
			if (stat.Mode() & os.ModeNamedPipe) == 0 {
				fmt.Println("[!] You should pass something to stdin or use the input-file option.")
				os.Exit(1)
			}
		}

		// Output option
		if outputOption != "text" && outputOption != "grep" && outputOption != "short" {
			outputOption = "text"
		}

	},
	Run: func(cmd *cobra.Command, args []string) {
		gopscannetwork.RunScanNetwork(reader, tcpScanOption, udpScanOption, portOption, onlyOpenOption, concurrencyOption, outputOption)
	},
}

func init() {
	scanNetworkCmd.Flags().StringVarP(&inputFileOption, "input-file", "i", "", "Input file with the IP addresses to scan. If no file is passed, then the stdin will be taken.")
	scanNetworkCmd.Flags().BoolVarP(&tcpScanOption, "tcp", "", false, "Scan with the TCP protocol.")
	scanNetworkCmd.Flags().BoolVarP(&udpScanOption, "udp", "", false, "[WIP] Scan with the UDP protocol.")
	scanNetworkCmd.Flags().StringVarP(&portOption, "port", "p", "", "Ports to scan. Can be either : \n\t- X,Y,Z \n\t- X-Y\n\t- X-Y,Z\nFamily of ports can also be passed : \n\t- http\n\t- ssh\n\t- mail\n\t- ...\nOptions can be combined, example :\n\t- 22,http,445,8080-8088")
	scanNetworkCmd.Flags().BoolVarP(&onlyOpenOption, "open", "", false, "Display only open ports.")
	scanNetworkCmd.Flags().IntVarP(&concurrencyOption, "concurrency", "t", 5000, "Number of threads used to take to scan.")
	scanNetworkCmd.Flags().StringVarP(&outputOption, "output", "o", "text", "Display result format as :\n\t- text\n\t- grep\n\t- short\nOption text will display a human-readable outpit.\nOption grep will displays the following format for each host :\n\tip,protcol,port,status.\nOption short will activate the flag --open and will display value for each open port as :\n\tip:port")
}
