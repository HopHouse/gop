package cmd

import (
	"github.com/spf13/cobra"
)

// serverCmd represents the serve command
var webCmd = &cobra.Command{
	Use: "web",
}

func init() {
	webCmd.AddCommand(crawlerCmd)
	webCmd.AddCommand(staticCrawlerCmd)
	webCmd.AddCommand(screenshotCmd)
	webCmd.AddCommand(visitCmd)
}
