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
	"github.com/hophouse/gop/gopOsintEmailGen"
	"github.com/hophouse/gop/utils/logger"
	"github.com/spf13/cobra"
)

var emailGenDomainOption string

// gopstaticcrawler represents the active command
var osintEmailGenCmd = &cobra.Command{
	Use:   "email",
	Short: "Generate email based on input data. It will create all the possible variations based on the allowed delimiters.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		logger.NewLoggerStdout()
	},
	Run: func(cmd *cobra.Command, args []string) {
		gopOsintEmailGen.RunEmailGen(firstNameOption, surnameOption, emailGenDomainOption, delimitersOption)
	},
}

func init() {
	osintEmailGenCmd.Flags().StringVarP(&firstNameOption, "firstname", "f", "", "First name.")
	_ = osintEmailGenCmd.MarkFlagRequired("firstname")
	osintEmailGenCmd.Flags().StringVarP(&surnameOption, "surname", "s", "", "Surname.")
	_ = osintEmailGenCmd.MarkFlagRequired("surname")
	osintEmailGenCmd.Flags().StringVarP(&emailGenDomainOption, "domain", "d", "", "Domain used into the email address.")
	_ = osintEmailGenCmd.MarkFlagRequired("domain")
}
