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
	gopirc "github.com/hophouse/gop/gopIrc"
	"github.com/hophouse/gop/utils"
	"github.com/spf13/cobra"
)

var (
	usernameOption string
)

var userNameOptionDefaultValue string = "anonymous"

// ircCmd represents the host command
var ircCmd = &cobra.Command{
	Use:   "irc",
	Short: "IRC module to chat.",
	Long:  "IRC module to chat.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		utils.NewLoggerStdout()

	},
}

// ircServerCmd represents the host command
var ircServerCmd = &cobra.Command{
	Use:   "server",
	Short: "IRC server module to chat.",
	Long:  "IRC server module to chat.",
	Run: func(cmd *cobra.Command, args []string) {
		gopirc.RunServerIRC(hostOption, portOption, usernameOption)
	},
}

// ircServerCmd represents the host command
var ircClientCmd = &cobra.Command{
	Use:   "client",
	Short: "IRC client module to chat.",
	Long:  "IRC client module to chat.",
	Run: func(cmd *cobra.Command, args []string) {
		gopirc.RunClientIRC(hostOption, portOption, usernameOption)
	},
}

func init() {
	ircCmd.AddCommand(ircServerCmd)
	ircCmd.AddCommand(ircClientCmd)

	ircCmd.PersistentFlags().StringVarP(&hostOption, "Host", "H", "127.0.0.1", "Define the irc server/client host.")
	ircCmd.PersistentFlags().StringVarP(&portOption, "Port", "P", "1337", "Define the irc server/client port.")
	ircCmd.PersistentFlags().StringVarP(&usernameOption, "username", "u", userNameOptionDefaultValue, "Username of the user.")
}
