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
	gopproxy "github.com/hophouse/gop/gopProxy"
	"github.com/spf13/cobra"
)

var (
	caFileOption        string
	caPrivKeyFileOption string
	verboseOption       bool
	interceptOption     bool
)

// proxyCmd represents the proxy command
var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Set up a proxy to use",
	Long:  "Set up a proxy to use",
	Run: func(cmd *cobra.Command, args []string) {
		options := gopproxy.InitOptions(hostOption, portOption, verboseOption, interceptOption, caFileOption, caPrivKeyFileOption)
		gopproxy.RunProxyCmd(&options)
	},
}

func init() {
	rootCmd.AddCommand(proxyCmd)

	proxyCmd.PersistentFlags().StringVarP(&hostOption, "Host", "H", "127.0.0.1", "Define the proxy host.")
	proxyCmd.PersistentFlags().StringVarP(&portOption, "Port", "P", "8080", "Define the proxy port.")
	proxyCmd.PersistentFlags().StringVarP(&caFileOption, "ca-file", "", "", "Certificate Authority certficate file.")
	proxyCmd.PersistentFlags().StringVarP(&caPrivKeyFileOption, "capriv-file", "", "", "Certificate Authority Private Key file.")
	proxyCmd.PersistentFlags().BoolVarP(&verboseOption, "verbose", "v", false, "Display more information about packets.")
	proxyCmd.PersistentFlags().BoolVarP(&reportOption, "intercept", "i", false, "Intercept traffic.")
}
