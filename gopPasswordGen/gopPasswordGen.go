package gopPasswordGen

import (
	"fmt"
	"os"
	"strings"

	"github.com/hophouse/gop/utils/logger"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var generatedWordlist []string

// RunEmailGen will create all the variations of email based on the inputed data.
func RunPasswordGen(wordlist []string, delimiters []string, minYear int, maxYear int, outFile string, stdinOption bool) {
	// Initial phase to add base words
	generatedWordlist = wordlist

	// Lower and uppercae words
	for _, word := range generatedWordlist {
		generatedWordlist = append(generatedWordlist, strings.ToLower(word))
		generatedWordlist = append(generatedWordlist, strings.ToUpper(word))
	}

	// Keep only unique values
	generatedWordlist = unique(generatedWordlist)

	// Capitalize words
	caser := cases.Title(language.French)
	for _, word := range generatedWordlist {
		generatedWordlist = append(generatedWordlist, caser.String(word))
	}

	// Keep only unique values
	generatedWordlist = unique(generatedWordlist)

	// Concatenante : [wordlist][delimiter][wordlist]
	// computedWordlist := []string{}
	// for _, word1 := range generatedWordlist {
	// 	for _, word2 := range generatedWordlist {
	// 		computed := fmt.Sprintf("%s%s", word1, word2)
	// 		computedWordlist = append(computedWordlist, computed)
	// 		for _, delimiter := range delimiters {
	// 			computed := fmt.Sprintf("%s%s%s", word1, delimiter, word2)
	// 			computedWordlist = append(computedWordlist, computed)
	// 		}
	// 	}
	// }
	// generatedWordlist = append(generatedWordlist, computedWordlist...)

	// Concatenante : [wordlist][delimiter][wordlist]
	computedWordlist := []string{}
	for _, word1 := range generatedWordlist {
		for _, word2 := range generatedWordlist {
			computed := fmt.Sprintf("%s%s", word1, word2)
			computedWordlist = append(computedWordlist, computed)
			for i := 0; i < 100; i++ {
				computed := fmt.Sprintf("%s%2d%s", word1, i, word2)
				computedWordlist = append(computedWordlist, computed)
			}
		}
	}
	generatedWordlist = append(generatedWordlist, computedWordlist...)

	// Keep only unique values
	generatedWordlist = unique(generatedWordlist)

	// Add years before and after each word : [delimiter][wordlist] & [wordlist][delimiter]
	computedWordlist = []string{}
	for _, word := range generatedWordlist {
		for i := minYear; i <= maxYear; i++ {
			var computed string

			// Add year before
			computed = fmt.Sprintf("%d%s", i, word)
			computedWordlist = append(computedWordlist, computed)

			// Add year after
			computed = fmt.Sprintf("%s%d", word, i)
			computedWordlist = append(computedWordlist, computed)

			// Add year before and after with delimiters
			for _, delimiter := range delimiters {
				// Add year before
				computed = fmt.Sprintf("%d%s%s", i, delimiter, word)
				computedWordlist = append(computedWordlist, computed)

				// Add year after
				computed = fmt.Sprintf("%s%s%d", word, delimiter, i)
				computedWordlist = append(computedWordlist, computed)
			}
		}
	}
	generatedWordlist = append(generatedWordlist, computedWordlist...)

	// Keep only unique values
	generatedWordlist = unique(generatedWordlist)

	/*
		// Add special chars before and after each word : [delimiter][wordlist] & [wordlist][delimiter]
		computedWordlist = []string{}
		for _, word := range generatedWordlist {
			for _, delimiter := range delimiters {
				computed := fmt.Sprintf("%s%s", delimiter, word)
				computedWordlist = append(computedWordlist, computed)
				computed = fmt.Sprintf("%s%s", word, delimiter)
				computedWordlist = append(computedWordlist, computed)
			}
		}
		generatedWordlist = append(generatedWordlist, computedWordlist...)
	*/

	// Keep only unique values
	generatedWordlist = unique(generatedWordlist)

	if stdinOption {
		for _, word := range generatedWordlist {
			logger.Println(word)
		}
	} else {
		f, err := os.OpenFile(outFile, os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			logger.Fatal("Error opening file to write generated password.")
		}
		defer f.Close()

		for _, word := range generatedWordlist {
			_, err := f.WriteString(word + "\n")
			if err != nil {
				logger.Println("Error writing in file.")
			}
		}
	}
}

func unique(slice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
