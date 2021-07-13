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

	gopvisit "github.com/hophouse/gop/gopVisit"
	"github.com/spf13/cobra"
)

var (
	burpOption bool
)

// visitCmd represents the visit command
var visitCmd = &cobra.Command{
	Use:   "visit",
	Short: "Visit supplied URLs.",
	Long:  "Visit supplied URLs.",
	PreRun: func(cmd *cobra.Command, args []string) {
		var err error

		// Parse options
		reader = os.Stdin

		stat, err := reader.Stat()
		if err != nil {
			fmt.Errorf("Error found in stdin:%s", err)
		}

		if inputFileOption != "" {
			if (stat.Mode() & os.ModeNamedPipe) != 0 {
				fmt.Println("[!] Cannot use stdin and input-file at the same time.")
				os.Exit(2)
			}
			reader, err = os.Open(inputFileOption)
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
	},
	Run: func(cmd *cobra.Command, args []string) {
		if proxyOption == "" && burpOption == true {
			proxyOption = "http://127.0.0.1:8080"
		}
		gopvisit.RunVisitCmd(reader, proxyOption)
	},
}

func init() {
	rootCmd.AddCommand(visitCmd)

	visitCmd.Flags().StringVarP(&inputFileOption, "input-file", "i", "", "Use the specified cookie.")
	visitCmd.Flags().StringVarP(&proxyOption, "proxy", "p", "", "Use this proxy to visit the pages.")
	visitCmd.Flags().BoolVarP(&burpOption, "burp", "", false, "Set the proxy directly to the default address and port of burp.")
}
