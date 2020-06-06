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
    "path/filepath"

	"github.com/spf13/cobra"
	"github.com/hophouse/gop/utils"
	"github.com/hophouse/gop/gopScreen"
)

var (
	timeoutOption int
	concurrencyOption int
)

// screenCmd represents the screen command
var screenCmd = &cobra.Command{
	Use:   "screenshot",
	Short: "Take screenshots of the supplied URLs. The program will take the stdin if no input file is passed as argument.",
	Long: "Take screenshots of the supplied URLs. The program will take the stdin if no input file is passed as argument.",
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
            
			reader, err = os.Open(filepath.Join("..", inputFileOption))
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

		// Create a specific log file for the screenshots
		utils.CreateOutputDir("screenshot")
	},
	Run: func(cmd *cobra.Command, args []string) {
		gopscreen.RunScreenCmd(reader, proxyOption, concurrencyOption, timeoutOption, delayOption)
	},
}

func init() {
	rootCmd.AddCommand(screenCmd)

	screenCmd.PersistentFlags().StringVarP(&inputFileOption ,"input-file", "i", "", "Use the specified cookie.")
	screenCmd.PersistentFlags().StringVarP(&proxyOption ,"proxy", "p", "", "Use this proxy to visit the pages.")
	screenCmd.PersistentFlags().IntVarP(&delayOption ,"delay", "", 0, "Use this delay in seconds between requests.")
	screenCmd.PersistentFlags().IntVarP(&concurrencyOption ,"concurrency", "t", 10, "Thread used to take screenshot.")
}
