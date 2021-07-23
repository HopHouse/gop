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
	goptunnel "github.com/hophouse/gop/gopTunnel"
	"github.com/hophouse/gop/utils"
	"github.com/spf13/cobra"
)

var (
	socketOption string
	tunnelOption string
)

// socksCmd represents the host command
var tunnelCmd = &cobra.Command{
	Use:   "tunnel",
	Short: ".",
	Long:  ".",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		utils.NewLoggerStdout()
	},
}

var tunnelServerCmd = &cobra.Command{
	Use:   "server",
	Short: ".",
	Long:  ".",
	Run: func(cmd *cobra.Command, args []string) {
		goptunnel.RunServer(tunnelOption, socketOption, modeOption)
	},
}

var tunnelClientCmd = &cobra.Command{
	Use:   "client",
	Short: ".",
	Long:  ".",
	Run: func(cmd *cobra.Command, args []string) {
		goptunnel.RunClient(tunnelOption, socketOption, modeOption)
	},
}

func init() {
	rootCmd.AddCommand(tunnelCmd)

	tunnelCmd.AddCommand(tunnelServerCmd)
	tunnelCmd.AddCommand(tunnelClientCmd)
	tunnelCmd.PersistentFlags().StringVarP(&modeOption, "mode", "m", "", "Choose which mode use : send, listen, or socks5.")
	tunnelCmd.PersistentFlags().StringVar(&tunnelOption, "tunnel", "127.0.0.1:6666", "Address where a socket will be open waiting for connexions to establish a tunnel.")
	tunnelCmd.PersistentFlags().StringVar(&socketOption, "redirect", "127.0.0.1:4444", "Address where the traffic will be redirected.")
}
