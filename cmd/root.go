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

	"github.com/hophouse/gop/utils"
	"github.com/spf13/cobra"
)

var (
	LogFile *os.File
)

var (
	proxyOption         string
	directoryNameOption string
	logFileNameOption   string
	directoryName       string
	inputFileOption     string
	stdinOption         bool
	reader              *os.File
	hostOption          string
	portOption          string
	concurrencyOption   int
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gop",
	Short: "GOP provides a toolbox to do pentest tasks.",
	Long:  "GOP provide a help performing some pentest tasks.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// If an output directory is specified, check if it exists and then move to it
		if directoryNameOption != "" {
			if _, err := os.Stat(directoryNameOption); os.IsNotExist(err) {
				utils.Log.Fatal("Specified directory by the option '-d' or '--directory', do not exists on the system.")
			}
			os.Chdir(directoryNameOption)
		} else {
			utils.CreateOutputDir(directoryNameOption)
		}

		// Init the logger and let close it later
		utils.NewLogger(LogFile, logFileNameOption)
		//defer utils.CloseLogger(LogFile)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&logFileNameOption, "logfile", "l", "logs.txt", "Set a custom log file.")
	rootCmd.PersistentFlags().StringVarP(&directoryNameOption, "output-directory", "D", "", "Use the following directory to output results.")
}
