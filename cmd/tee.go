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
	"log"

	goptee "github.com/hophouse/gop/gopTee"
	"github.com/hophouse/gop/utils"
	"github.com/spf13/cobra"
)

// teeCmd represents the proxy command
var teeCmd = &cobra.Command{
	Use:   "[WIP]tee",
	Short: "Act as the unix tee command but also display the executed command.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		utils.NewLoggerStdout()
	},
	Run: func(cmd *cobra.Command, args []string) {
		err := goptee.RunTeeCmd(outFileOption, cmdOption)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	teeCmd.PersistentFlags().StringVarP(&outFileOption, "output", "o", "", "Output of the tee command.")
	teeCmd.PersistentFlags().StringVarP(&cmdOption, "command", "c", "", "Command to execute.")
}
