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

	gopdynamiccrawler "github.com/hophouse/gop/gopDynamicCrawler"
	"github.com/spf13/cobra"
)

// crawlerCmd represents the crawler command
var crawlerCmd = &cobra.Command{
	Use:   "crawler",
	Short: "Crawler command to crawl recursively or not a domain or a website.",
	Long:  "Crawler command to crawl recursively or not a domain or a website.",
	PreRun: func(cmd *cobra.Command, args []string) {
		gopdynamiccrawler.NewOptions(&UrlOption, LogFile, &reportOption, &recursiveOption, &screenshotOption, &cookieOption, &proxyOption, &delayOption, &concurrencyOption)

		// Screenshots directory and HTML page
		if screenshotOption == true {
			if _, err := os.Stat("screenshots"); os.IsNotExist(err) {
				os.Mkdir("screenshots", 0600)
			}
		}

		// Print Options
		gopdynamiccrawler.PrintOptions(&gopdynamiccrawler.GoCrawlerOptions)
	},
	Run: func(cmd *cobra.Command, args []string) {
		gopdynamiccrawler.RunCrawlerCmd()
	},
}

func init() {
	rootCmd.AddCommand(crawlerCmd)

	crawlerCmd.PersistentFlags().StringVarP(&UrlOption, "url", "u", "", "URL to test.")
	crawlerCmd.MarkPersistentFlagRequired("url")
	crawlerCmd.PersistentFlags().BoolVarP(&recursiveOption, "recursive", "r", false, "Crawl the website recursively.")
	crawlerCmd.PersistentFlags().BoolVarP(&screenshotOption, "screenshot", "s", false, "Take a screenshot on each visited link.")
	//crawlerCmd.PersistentFlags().StringVarP(&cookieOption, "cookie", "c", "", "Use the specified cookie.")
	crawlerCmd.PersistentFlags().StringVarP(&proxyOption, "proxy", "p", "", "Use the specified proxy.")
	crawlerCmd.PersistentFlags().IntVarP(&delayOption, "delay", "", 0, "Use this delay in seconds between each requests.")
	crawlerCmd.PersistentFlags().IntVarP(&concurrencyOption, "concurrency", "t", 10, "Thread used to take screenshot.")
}
