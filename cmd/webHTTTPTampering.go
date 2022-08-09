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

	gopwebtampering "github.com/hophouse/gop/gopWebTampering"
	"github.com/hophouse/gop/utils"
	"github.com/spf13/cobra"
)

var (
	webRequestFilenameOption string
	webRequestUrlOption      string
	webOnlyStatusCodeOption  []int
	webValidResourcesOption  []string
)

// headerTamperingCmd represents the active command
var headerTamperingCmd = &cobra.Command{
	Use:   "tamper",
	Short: "Couple of action to add/modify/delete headers and find abnormal behaviors. It will log every URL but only print matches.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.PersistentPreRun(cmd, args)

		if burpOption {
			gopwebtampering.Options.Proxy = "http://127.0.0.1:8080"
		} else {
			gopwebtampering.Options.Proxy = proxyOption
		}
	},
}

// headerTamperingCmd represents the active command
var allTamperingCmd = &cobra.Command{
	Use:   "all",
	Short: "Run all the tampering subcommands.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print("[+] Host header tampering\n")
		err := gopwebtampering.TamperHostHeader(webRequestFilenameOption)
		if err != nil {
			utils.Log.Fatalln(err)
		}
		fmt.Print("\n")

		fmt.Print("[+] Referer header tampering\n")
		err = gopwebtampering.TamperReferrerHeader(webRequestFilenameOption)
		if err != nil {
			utils.Log.Fatalln(err)
		}
		fmt.Print("\n")

		fmt.Print("[+] IP source tampering\n")
		err = gopwebtampering.TamperIPSource(webRequestFilenameOption)
		if err != nil {
			utils.Log.Fatalln(err)
		}
		fmt.Print("\n")
	},
}

// "Host" header tamper represents the active command
var hostHeaderTamperingCmd = &cobra.Command{
	Use:   "host",
	Short: "Tamper IP source through the HOST header and the reverse-proxy related headers.",
	Run: func(cmd *cobra.Command, args []string) {
		err := gopwebtampering.TamperHostHeader(webRequestFilenameOption)
		if err != nil {
			utils.Log.Fatalln(err)
		}
	},
}

// "Referer" header tamper represents the active command
var referrerHeaderTamperingCmd = &cobra.Command{
	Use:   "referrer",
	Short: "Tamper the \"Referer\" header.",
	Run: func(cmd *cobra.Command, args []string) {
		err := gopwebtampering.TamperReferrerHeader(webRequestFilenameOption)
		if err != nil {
			utils.Log.Fatalln(err)
		}
	},
}

// source IP tamper represents the active command
var sourceIPTamperingCmd = &cobra.Command{
	Use:   "sourceIP",
	Short: "Tamper IP source through the HOST header and the reverse-proxy related headers.",
	Run: func(cmd *cobra.Command, args []string) {
		err := gopwebtampering.TamperIPSource(webRequestFilenameOption)
		if err != nil {
			utils.Log.Fatalln(err)
		}
	},
}

// Nginx off-by-slash fail tamper represents the active command
var NginxOffySlashCmd = &cobra.Command{
	Use:   "nginxOffBySlash",
	Short: "Tamper the URL/request in order to discover an Nginx off-by-slash vulnerability.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.PersistentPreRun(cmd, args)

		if webRequestUrlOption != "" && webRequestFilenameOption != "" {
			fmt.Print("Could not use an URL and a request file at the same time.\n")
			utils.Log.Fatal("Could not use an URL and a request file at the same time.\n")
		}

		if burpOption {
			gopwebtampering.Options.Proxy = "http://127.0.0.1:8080"
		} else {
			gopwebtampering.Options.Proxy = proxyOption
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		err := gopwebtampering.NginxOffBySlash(webRequestFilenameOption, webRequestUrlOption, webValidResourcesOption, webOnlyStatusCodeOption)
		if err != nil {
			utils.Log.Fatalln(err)
		}
	},
}

func init() {
	headerTamperingCmd.PersistentFlags().StringVarP(&webRequestFilenameOption, "request", "r", "", "File where the request to tamper is.")
	headerTamperingCmd.MarkPersistentFlagRequired("request")

	headerTamperingCmd.PersistentFlags().StringVarP(&proxyOption, "proxy", "p", "", "Use the specified proxy.")
	headerTamperingCmd.PersistentFlags().BoolVarP(&burpOption, "burp", "", false, "Set burp as proxy with default configuration.")

	headerTamperingCmd.AddCommand(allTamperingCmd)
	headerTamperingCmd.AddCommand(hostHeaderTamperingCmd)
	headerTamperingCmd.AddCommand(referrerHeaderTamperingCmd)
	headerTamperingCmd.AddCommand(sourceIPTamperingCmd)

	nginxOffBySlashValidResourcesDefaultValues := []string{
		"etc/passwd",
		"var/log/nginx/access.log",
		"var/log/lastlog",
		"log/nginx/access.log",
		"log/lastlog",
		".git/HEAD",
		".env",
		".htaccess",
		"robots.txt",
		"nginx.conf",
		"README.md",
		"index.html",
		"html/index.html",
		"public/index.html",
		"settings.php",
		"config/settings.php",
		"config/index.php",
	}

	NginxOffySlashCmd.PersistentFlags().StringVarP(&webRequestFilenameOption, "request", "r", "", "File where the request to tamper is.")
	NginxOffySlashCmd.PersistentFlags().StringVarP(&webRequestUrlOption, "url", "u", "", "URL where resource needs to be tamperd.")
	NginxOffySlashCmd.PersistentFlags().StringVarP(&proxyOption, "proxy", "p", "", "Use the specified http proxy (ex: http://127.0.0.1:8080).")
	NginxOffySlashCmd.PersistentFlags().BoolVarP(&burpOption, "burp", "", false, "Set burp as proxy with default configuration (http://127.0.0.1:8080).")
	NginxOffySlashCmd.PersistentFlags().StringSliceVarP(&webValidResourcesOption, "path", "", nginxOffBySlashValidResourcesDefaultValues, "Valid resources to request in the form : 'static/app.js'. By default a pre-compiled list was created")
	NginxOffySlashCmd.PersistentFlags().IntSliceVarP(&webOnlyStatusCodeOption, "show", "", []int{}, "Only show the specified status codes.")
}
