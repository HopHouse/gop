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
	"debug/pe"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/hophouse/gop/gopBin"
	"github.com/hophouse/gop/utils/logger"
	"github.com/spf13/cobra"
)

var (
	minCaveSizeOption int
	byteFileOption    string
	offsetOption      int64
)

// binCmd represents the bin command
var binCmd = &cobra.Command{
	Use: "bin",
}

var binEntropyCmd = &cobra.Command{
	Use:   "entropy",
	Short: "Get the entropy of a file",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		logger.NewLoggerNull()
	},
	Run: func(cmd *cobra.Command, args []string) {
		f, err := gopBin.GetFileEntropyHandle(inputFileOption)
		if err != nil {
			log.Fatalln(err)
		}
		f.ParseBytes(inputFileOption)
		f.PrintEntropy()
	},
}

var binCaveCmd = &cobra.Command{
	Use:   "cave",
	Short: "Find cave in sections of a binary",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		logger.NewLoggerNull()
	},
	Run: func(cmd *cobra.Command, args []string) {
		err := gopBin.GetPECaves(inputFileOption, minCaveSizeOption)
		if err != nil {
			log.Fatalln(err)
		}

	},
}

var binWriteBytesCmd = &cobra.Command{
	Use:   "write",
	Short: "Write bytes at offset in a file",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		logger.NewLoggerNull()
	},
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Printf("[+] File name     : %s\n", inputFileOption)
		fmt.Printf("[+] Offset        : %d\n", offsetOption)
		fmt.Printf("[+] Bytefile name : %s\n", byteFileOption)

		fOpen, err := os.OpenFile(inputFileOption, os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println(err)
			return

		}
		defer fOpen.Close()

		fData, err := os.Open(byteFileOption)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer fData.Close()

		data, err := io.ReadAll(fData)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("[+] Bytefile size : %d\n", len(data))

		n, err := fOpen.WriteAt(data, offsetOption)
		if err != nil {
			fmt.Println(err)
			return
		}

		if n != len(data) {
			fmt.Printf("[!] Written %d bytes but expected %d bytes\n", n, len(data))
			return
		}

		fmt.Printf("[+] Successfully written %d bytes\n", n)
	},
}

var binPEInfoCmd = &cobra.Command{
	Use:   "PE info",
	Short: "Get information about the PE",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		logger.NewLoggerNull()
	},
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Printf("[+] PE name : %s\n", inputFileOption)

		PE, err := pe.Open(inputFileOption)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer PE.Close()

		fmt.Printf("[+] Sections :\n")
		for _, section := range PE.Sections {
			fmt.Printf("\t%s - offset : 0x%08x (%d) - size : %d\n", section.Name, section.Size, section.Size, section.Offset)
		}

		fmt.Printf("[+] Sections details:\n")
		for _, section := range PE.Sections {
			fmt.Printf("\t%s :\n", section.Name)
			fmt.Printf("\t\tVirtual address : 0x%016x\n", section.VirtualAddress)
			fmt.Printf("\t\tVirtual size    : %d\n", section.VirtualSize)
		}

		libs, err := PE.ImportedLibraries()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("[+] Imported Library :\n")
		for i, lib := range libs {
			fmt.Printf("\t%d. %s\n", i, lib)

		}

		symbols, err := PE.ImportedSymbols()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("[+] Imported Symbols :\n")
		for i, symbol := range symbols {
			fmt.Printf("\t%d. %s\n", i, symbol)
		}

		fmt.Printf("[+] File headers : %#v\n", PE.FileHeader)

		switch PE.FileHeader.Machine {
		case 0x014c:
			fmt.Printf("[+] Entry point address: %#v\n", PE.OptionalHeader.(*pe.OptionalHeader32).AddressOfEntryPoint)
		case 0x0200:
			fmt.Printf("[+] Entry point address : Could not handle Intel Itanium\n")
		case 0x8664:
			fmt.Printf("[+] Entry point address: %#v\n", PE.OptionalHeader.(*pe.OptionalHeader64).AddressOfEntryPoint)
		}
	},
}

func init() {
	binCmd.AddCommand(binEntropyCmd)
	binCmd.AddCommand(binCaveCmd)
	binCmd.AddCommand(binWriteBytesCmd)
	binCmd.AddCommand(binPEInfoCmd)

	binCmd.PersistentFlags().StringVarP(&inputFileOption, "file", "f", "", "File to compare the stdin with.")
	binCmd.MarkFlagRequired("file")

	binCaveCmd.Flags().IntVarP(&minCaveSizeOption, "min-cave-size", "s", 128, "Minimum size to looke for caves in the PE.")

	binWriteBytesCmd.Flags().StringVarP(&inputFileOption, "file", "f", "", "File to compare the stdin with.")
	binWriteBytesCmd.MarkFlagRequired("file")
	binWriteBytesCmd.Flags().StringVarP(&byteFileOption, "byte-file", "b", "", "File with bytes to write into.")
	binWriteBytesCmd.MarkFlagRequired("byte-file")
	binWriteBytesCmd.Flags().Int64VarP(&offsetOption, "offset", "", 0, "Offset where to write bytes.")
	binWriteBytesCmd.MarkFlagRequired("offset")
}
