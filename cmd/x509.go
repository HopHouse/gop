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
	"bufio"
	"os"

	"github.com/hophouse/gop/gopX509"
	"github.com/hophouse/gop/utils/logger"
	"github.com/spf13/cobra"
)

var (
	addressOption []string
)

// x509Cmd represents the active command
var x509Cmd = &cobra.Command{
	Use: "x509",
}

// headerTamperingCmd represents the active command
var x509NamesCmd = &cobra.Command{
	Use:   "names",
	Short: "Extract CN (Common Name) and SAN (Subject Alternative Name) of a certificate",
	PreRun: func(cmd *cobra.Command, args []string) {
		var err error

		// Parse options
		reader = os.Stdin

		stat, err := reader.Stat()
		if err != nil {
			logger.Println("Error found in stdin:", err)
			os.Exit(2)
		}

		if len(addressOption) > 0 {
			if (stat.Mode() & os.ModeNamedPipe) != 0 {
				logger.Println("[!] Cannot use stdin and address at the same time.")
				os.Exit(2)
			}
			reader, err = os.Open(inputFileOption)
			if err != nil {
				panic(err)
			}
		} else {
			// Chech if there is something in stdin
			if (stat.Mode() & os.ModeNamedPipe) == 0 {
				logger.Println("[!] You should pass something to stdin or use the address option.")
				os.Exit(1)
			}
		}

		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			addressOption = append(addressOption, scanner.Text())
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		err := gopX509.RunX509Names(addressOption)
		if err != nil {
			logger.Fatal(err)
		}
	},
}

func init() {
	x509NamesCmd.PersistentFlags().StringSliceVarP(&addressOption, "address", "a", []string{}, "Address where to look for : Ex : 10.0.0.0:443.")

	x509Cmd.AddCommand(x509NamesCmd)
}
