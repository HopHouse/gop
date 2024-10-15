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
	"os"

	gophost "github.com/hophouse/gop/gopHost"
	"github.com/hophouse/gop/utils/logger"
	"github.com/spf13/cobra"
)

var petitPoucetOption bool
var freeOption bool

// hostCmd represents the host command
var hostCmd = &cobra.Command{
	Use:   "host",
	Short: "Resolve hostname to get the IP address.",
	Long:  "Resolve hostname to get the IP address.",
	PreRun: func(cmd *cobra.Command, args []string) {
		var err error

		// Parse options
		reader = os.Stdin

		stat, err := reader.Stat()
		if err != nil {
			logger.Println("Error found in stdin:", err)
			os.Exit(2)
		}

		if inputFileOption != "" {
			if (stat.Mode() & os.ModeNamedPipe) != 0 {
				logger.Println("[!] Cannot use stdin and input-file at the same time.")
				os.Exit(2)
			}
			reader, err = os.Open(inputFileOption)
			if err != nil {
				panic(err)
			}
		} else {
			// Chech if there is something in stdin
			if (stat.Mode() & os.ModeNamedPipe) == 0 {
				logger.Println("[!] You should pass something to stdin or use the input-file option.")
				os.Exit(1)
			}
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		gophost.RunHostCmd(reader, concurrencyOption, petitPoucetOption, freeOption)
	},
}

func init() {
	hostCmd.PersistentFlags().StringVarP(&inputFileOption, "input-file", "i", "", "Specify domain names to check host.")
	hostCmd.PersistentFlags().IntVarP(&concurrencyOption, "concurrency", "t", 10, "Number of thread used to check hosts.")
	hostCmd.PersistentFlags().BoolVarP(&petitPoucetOption, "petit-poucet", "", false, "Activate the petit poucet option which display all the CNAMES during the resolve process.")
	hostCmd.PersistentFlags().BoolVarP(&freeOption, "free", "", false, "[WIP] Check if the domain is avaialable for sale.")
}
