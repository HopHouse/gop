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
	"log"
	"os"

	"github.com/hophouse/gop/utils/logger"
	"github.com/spf13/cobra"
)

var (
	LogFile             *os.File
	CurrentDirectory    string
	CurrentLogDirectory string
)

var (
	proxyOption         string
	directoryNameOption string
	logFileNameOption   string
	noLogOption         bool
	inputFileOption     string
	stdinOption         bool
	reader              *os.File
	hostOption          string
	portOption          string
	concurrencyOption   int
	UrlOption           string
	recursiveOption     bool
	screenshotOption    bool
	cookieOption        string
	delayOption         int
	reportOption        bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gop",
	Short: "GOP provides a toolbox to do pentest tasks.",
	Long:  "GOP provides a toolbox to do pentest tasks.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		var err error

		if noLogOption {
			logger.NewLoggerNull()
			return
		}

		CurrentDirectory, err = os.Getwd()
		if err != nil {
			log.Fatal(err)
		}

		directoryNameOption = logger.CreateOutputDir(directoryNameOption, cmd.Use)

		// Move to the new generated folder
		os.Chdir(directoryNameOption)
		mydir, err := os.Getwd()
		if err != nil {
			fmt.Printf("[-] Error when created log dir %s", mydir)
		}

		logger.NewLoggerDateTime("logs.txt")
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logger.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&logFileNameOption, "logfile", "", "logs.txt", "Set a custom log file.")
	rootCmd.PersistentFlags().BoolVarP(&noLogOption, "no-log", "", false, "Do not create a log file.")
	rootCmd.PersistentFlags().StringVarP(&directoryNameOption, "output-directory", "", "", "Use the following directory to output results.")

	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(hostCmd)
	rootCmd.AddCommand(ircCmd)
	rootCmd.AddCommand(killSwitchCmd)
	rootCmd.AddCommand(proxyCmd)
	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(scheduleCmd)
	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(shellCmd)
	rootCmd.AddCommand(teeCmd)
	rootCmd.AddCommand(tunnelCmd)
	rootCmd.AddCommand(webCmd)
	rootCmd.AddCommand(crawlerCmd)
	rootCmd.AddCommand(screenshotCmd)
	rootCmd.AddCommand(staticCrawlerCmd)
	rootCmd.AddCommand(stationCmd)
	rootCmd.AddCommand(visitCmd)
	rootCmd.AddCommand(x509Cmd)
	rootCmd.AddCommand(fileDiffCmd)
	rootCmd.AddCommand(encodeCmd)
}
