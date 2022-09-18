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
	"github.com/hophouse/gop/gopStation"
	"github.com/hophouse/gop/utils"
	"github.com/spf13/cobra"
)

var (
	tcpOption string
	sslOption string
)

// screenCmd represents the screen command
var stationCmd = &cobra.Command{
	Use:     "station",
	Short:   "Use it as a station to manage and retrieve simple reverse shell in plain TCP.",
	Version: "0.1",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.PersistentPreRun(cmd, args)
		utils.NewLoggerNull()
	},
	Run: func(cmd *cobra.Command, args []string) {
		gopStation.RunServerCmd(tcpOption, sslOption)
	},
}

func init() {
	stationCmd.PersistentFlags().StringVarP(&tcpOption, "tcp", "t", "127.0.0.1:8080", "Define the tcp listening address.")
	stationCmd.PersistentFlags().StringVarP(&sslOption, "ssl", "s", "127.0.0.1:8443", "Define the ssl listening address.")
}
