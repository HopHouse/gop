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
	"io/ioutil"
	"net/http"
	"os/exec"
	"strings"

	"github.com/hophouse/gop/utils/logger"
	"github.com/spf13/cobra"
)

var (
	messageOption string
)

// ircCmd represents the host command
var killSwitchCmd = &cobra.Command{
	Use: "kill-switch",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		logger.NewLoggerStdout()
	},
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := http.Get(UrlOption)
		if err != nil {
			logger.Println(err)
			return
		}

		if resp.StatusCode != 200 {
			if err != nil {
				logger.Println(err)
				return
			}
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logger.Println(err)
			return
		}
		defer resp.Body.Close()

		if strings.Contains(string(body), messageOption) {
			command := strings.Split(cmdOption, " ")
			cmd := exec.Command(command[0], command[1:]...)
			err := cmd.Run()
			if err != nil {
				logger.Println(err)
				return
			}
		}

	},
}

func init() {
	killSwitchCmd.PersistentFlags().StringVarP(&UrlOption, "url", "u", "", "Url to fetch")
	killSwitchCmd.PersistentFlags().StringVarP(&messageOption, "message", "m", "", "Message to look for.")
	killSwitchCmd.PersistentFlags().StringVarP(&cmdOption, "cmd", "c", "", "Command to execute if condition is met")
}
