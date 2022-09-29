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
	"path/filepath"

	gopscreen "github.com/hophouse/gop/gopScreen"
	"github.com/hophouse/gop/utils/logger"
	"github.com/spf13/cobra"
)

var (
	timeoutOption    int
	doNotAppendSlash bool
)

// screenCmd represents the screen command
var screenshotCmd = &cobra.Command{
	Use:     "screenshot",
	Short:   "Take screenshots of the supplied URLs. The program will take the stdin if no input file is passed as argument.",
	Long:    "Take screenshots of the supplied URLs. The program will take the stdin if no input file is passed as argument.",
	Version: "0.2",
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

			reader, err = os.Open(filepath.Join("..", inputFileOption))
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
		gopscreen.RunScreenCmd(reader, proxyOption, concurrencyOption, timeoutOption, delayOption, cookieOption, doNotAppendSlash)
	},
}

func init() {
	screenshotCmd.Flags().StringVarP(&inputFileOption, "input-file", "i", "", "Use the specified file as input.")
	screenshotCmd.Flags().StringVarP(&proxyOption, "proxy", "p", "", "Use this proxy to visit the pages.")
	screenshotCmd.Flags().IntVarP(&delayOption, "delay", "", 0, "Use this delay in seconds between requests.")
	screenshotCmd.Flags().IntVarP(&timeoutOption, "timeout", "", 60, "Timeout before the chrome context is canceled.")
	screenshotCmd.Flags().IntVarP(&concurrencyOption, "concurrency", "t", 2, "Thread used to take screenshot.")
	screenshotCmd.Flags().StringVarP(&cookieOption, "cookie", "c", "", "Use the specified cookie.")
	screenshotCmd.Flags().BoolVarP(&doNotAppendSlash, "do-not-append-slash", "", false, "Do not append slash (\"/\") at the end of the URL if there is not.")
}
