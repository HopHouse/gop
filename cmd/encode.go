/*
Copyright © 2023 Hophouse <contact@hophouse.fr>

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
	"encoding/base32"
	"encoding/base64"
	"fmt"

	"github.com/spf13/cobra"
)

var decodeOption bool

// encodeCmd represents the crawler command
var encodeCmd = &cobra.Command{
	Use: "encode",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var encodeB64Cmd = &cobra.Command{
	Use: "b64",
	Run: func(cmd *cobra.Command, args []string) {
		if !decodeOption {
			fmt.Println(base64.StdEncoding.EncodeToString([]byte(messageOption)))
		} else {
			decoded, err := base64.StdEncoding.DecodeString(messageOption)
			if err != nil {
				fmt.Println("Error :", err)
				return
			}
			fmt.Println(string(decoded))
		}
	},
}

var encodeB32Cmd = &cobra.Command{
	Use: "b32",
	Run: func(cmd *cobra.Command, args []string) {
		if !decodeOption {
			fmt.Println(base32.StdEncoding.EncodeToString([]byte(messageOption)))
		} else {
			decoded, err := base32.StdEncoding.DecodeString(messageOption)
			if err != nil {
				fmt.Println("Error :", err)
				return
			}
			fmt.Println(string(decoded))
		}
	},
}

var encodeB64B32Cmd = &cobra.Command{
	Use: "b64b32",
	Run: func(cmd *cobra.Command, args []string) {
		if !decodeOption {
			fmt.Println(base32.StdEncoding.EncodeToString([]byte(base64.StdEncoding.EncodeToString([]byte(messageOption)))))
		} else {
			predecoded, err := base32.StdEncoding.DecodeString(messageOption)
			if err != nil {
				fmt.Println("Error :", err)
				return
			}

			decoded, err := base64.StdEncoding.DecodeString(string(predecoded))
			if err != nil {
				fmt.Println("Error :", err)
				return
			}

			fmt.Println(string(decoded))
		}
	},
}

func init() {
	encodeCmd.AddCommand(encodeB32Cmd)
	encodeCmd.AddCommand(encodeB64Cmd)
	encodeCmd.AddCommand(encodeB64B32Cmd)

	encodeCmd.PersistentFlags().StringVarP(&messageOption, "message", "m", "", "Message to encode or decode.")
	_ = encodeCmd.MarkPersistentFlagRequired("message")

	encodeCmd.PersistentFlags().BoolVarP(&decodeOption, "decode", "d", false, "Decode instead of encode.")
}
