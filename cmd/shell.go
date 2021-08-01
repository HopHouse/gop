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
	gopshell "github.com/hophouse/gop/gopShell"
	"github.com/hophouse/gop/utils"
	"github.com/spf13/cobra"
)

var (
	modeOption string
)

// shellCmd represents the proxy command
var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Set up a shell either reverse or bind.",
	Long:  "Set up a shell either reverse or bind.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		utils.NewLoggerStdout()
	},
	Run: func(cmd *cobra.Command, args []string) {
		gopshell.RunShellCmd(modeOption, hostOption, portOption)
	},
}

func init() {
	shellCmd.PersistentFlags().StringVarP(&hostOption, "Host", "H", "127.0.0.1", "Define the proxy host.")
	shellCmd.PersistentFlags().StringVarP(&portOption, "Port", "P", "8000", "Define the proxy port.")
	shellCmd.PersistentFlags().StringVarP(&modeOption, "Mode", "m", "reverse", "Define the mode where the shell is runned : blind or reverse.")
}
