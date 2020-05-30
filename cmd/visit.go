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
	"fmt"

	"github.com/spf13/cobra"
	"github.com/hophouse/gop/gopVisit"
)

// visitCmd represents the visit command
var visitCmd = &cobra.Command{
	Use:   "visit",
	Short: "Visit supplied URLs.",
	Long: "Visit supplied URLs.",
	PreRun: func(cmd *cobra.Command, args []string) {
		var err error
		reader = os.Stdin
		if (inputFileOption != "" && stdinOption == true) {
			fmt.Println("[!] Cannot use stdin and input-file at the same time.")
			os.Exit(2)
		}

		if (inputFileOption == "" && stdinOption == false) {
			fmt.Println("[!] Please at least specifiy the stdin or input-file option.")
			os.Exit(2)
		}

		if inputFileOption != "" {
			reader, err = os.Open(inputFileOption)
			if err != nil {
				panic(err)
			}
		}
		if stdinOption == true {
			stat, err := reader.Stat()
			if err != nil {
				fmt.Errorf("Error found in stdin:%s", err)
			}
			if (stat.Mode() & os.ModeNamedPipe) == 0 {
				fmt.Println("[!] You should pass something to stdin or use the input-file option.")
				os.Exit(2)
			}
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		gopvisit.RunVisitCmd(reader, proxyOption)
	},
}

func init() {
	rootCmd.AddCommand(visitCmd)

	visitCmd.PersistentFlags().StringVarP(&inputFileOption ,"input-file", "i", "", "Use the specified cookie.")
	visitCmd.PersistentFlags().BoolVarP(&stdinOption ,"stdin", "", false, "Value will be passed from stdin.")
	visitCmd.PersistentFlags().StringVarP(&proxyOption ,"proxy", "p", "", "Use this proxy to visit the pages.")
}
