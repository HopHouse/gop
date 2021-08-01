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
	"github.com/hophouse/gop/gopPasswordGen"
	"github.com/hophouse/gop/utils"
	"github.com/spf13/cobra"
)

var (
	baseWordlistOption []string
	wordlistOption     []string
	minYearOption      int
	maxYearOption      int
	outFileOption      string
)

// gopstaticcrawler represents the active command
var passwordGenCmd = &cobra.Command{
	Use:   "password",
	Short: "Generate password based on a given list of words.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		utils.NewLoggerStdout()
	},
	Run: func(cmd *cobra.Command, args []string) {
		wordlist := baseWordlistOption
		if len(wordlistOption) > 0 {
			wordlist = append(wordlist, wordlistOption...)
		}
		gopPasswordGen.RunPasswordGen(wordlist, delimitersOption, minYearOption, maxYearOption, outFileOption, stdinOption)
	},
}

func init() {
	Delimiters = []string{
		".",
		",",
		"-",
		"_",
		"#",
		"$",
		"%",
		"&",
		"*",
		"+",
		"/",
		"=",
		"!",
		"?",
		"^",
		"~",
	}

	baseWordlist := []string{
		"admin",
		"adm",
		"welcome",
		"test",
		"user",
		"pass",
		"password",
		"motdepasse",
	}

	passwordGenCmd.Flags().StringSliceVarP(&baseWordlistOption, "base-wordlist", "", baseWordlist, "baseWordlist.")
	passwordGenCmd.Flags().StringSliceVarP(&wordlistOption, "wordlist", "w", []string{}, "Word that will be append to the base-wordlist list.")
	passwordGenCmd.Flags().IntVarP(&minYearOption, "min-year", "", 2015, "min year.")
	passwordGenCmd.Flags().IntVarP(&maxYearOption, "max-year", "", 2022, "max year.")
	passwordGenCmd.Flags().StringVarP(&outFileOption, "outfile", "o", "gop_generate_passwords.txt", "File where the generated passwords are put.")
	passwordGenCmd.Flags().BoolVarP(&stdinOption, "stdin", "", false, "Display generated passwords in stdin instead of in a file.")
}
