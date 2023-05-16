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

	"github.com/hophouse/gop/utils/logger"
	"github.com/spf13/cobra"
)

// fileCmd represents the file command
var fileCmd = &cobra.Command{
	Use: "file",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
	},
}

// fileCmd represents the host command
var fileDiffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Compare stdin input to a file and return only the differences with this file",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		var err error

		logger.NewLoggerStdout()

		// Parse options
		reader = os.Stdin

		stat, err := reader.Stat()
		if err != nil {
			logger.Println("Error found in stdin:", err)
			os.Exit(2)
		}

		// Chech if there is something in stdin
		if (stat.Mode() & os.ModeNamedPipe) == 0 {
			logger.Println("[!] You should pass something to stdin or use the input-file option.")
			os.Exit(1)
		}

	},
	Run: func(cmd *cobra.Command, args []string) {
		known := make(map[string]bool)

		// Fill the list with known items
		f, err := os.Open(inputFileOption)
		if err != nil {
			logger.Fatalln(err)
		}
		defer f.Close()

		inputFileScanner := bufio.NewScanner(f)
		for inputFileScanner.Scan() {
			text := inputFileScanner.Text()
			known[text] = false
		}

		// Compare with stdin
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			text := scanner.Text()
			_, ok := known[text]
			if ok {
				known[text] = true
			} else {
				logger.Printf("+ %s\n", text)
			}
		}

		for key, value := range known {
			if !value {
				logger.Printf("- %s\n", key)
			}
		}
	},
}

func init() {
	fileDiffCmd.Flags().StringVarP(&inputFileOption, "file", "f", "", "File to compare the stdin with.")
	fileDiffCmd.MarkFlagRequired("file")
}
