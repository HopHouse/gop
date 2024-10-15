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
	"github.com/hophouse/gop/gopOsintUsernameGen"
	"github.com/hophouse/gop/utils/logger"
	"github.com/spf13/cobra"
)

var (
	firstNameOption string
	surnameOption   string
)

// gopstaticcrawler represents the active command
var osintUsernameGenCmd = &cobra.Command{
	Use:   "username",
	Short: "Generate username based on input data. It will create all the possible variations based on the allowed delimiters.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		logger.NewLoggerStdout()
	},
	Run: func(cmd *cobra.Command, args []string) {
		gopOsintUsernameGen.RunUsernameGen(firstNameOption, surnameOption, delimitersOption)
	},
}

func init() {
	osintUsernameGenCmd.Flags().StringVarP(&firstNameOption, "firstname", "f", "", "First name.")
	_ = osintUsernameGenCmd.MarkFlagRequired("firstname")
	osintUsernameGenCmd.Flags().StringVarP(&surnameOption, "surname", "s", "", "Surname.")
	_ = osintUsernameGenCmd.MarkFlagRequired("surname")
}
