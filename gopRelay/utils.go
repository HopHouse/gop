package gopRelay

import (
	"bytes"
	"fmt"
)

// Return :
// - 0 if equals
// - 1 if len(s1) > len(s2)
// - -1 if len(s1) < len(s2)
// - 2 if bitwise comparaison of s1 != s2
// TODO Mutualise with the one in the ntlm package
func CompareBytesSlices(s1, s2 []byte) (int, string) {
	status := 2
	str := bytes.NewBuffer([]byte{})

	if len(s1) > len(s2) {
		status = 1
		fmt.Fprintf(str, "\033[31mlength of s1 %d greater than length of s2 %d\n\033[0m", len(s1), len(s2))
	} else if len(s1) < len(s2) {
		status = -1
		fmt.Fprintf(str, "\033[31mlength of s1 %d lower than length of s2 %d\n\033[0m", len(s1), len(s2))
	} else if len(s1) == len(s2) {
		status = 0
		fmt.Fprintf(str, "\033[32mlength of s1 %d equals length of s2 %d\n\033[0m", len(s1), len(s2))
	}

	fmt.Fprintf(str, "\tg\tb\n")
	for i := 0; i < min(len(s1), len(s2)); i++ {
		if s1[i] == s2[i] {
			fmt.Fprintf(str, "%d\t%02x\t%02x\n", i, s1[i], s2[i])
		} else {
			status = 2
			fmt.Fprintf(str, "\033[31m%d\t%02x\t%02x\n\033[0m", i, s1[i], s2[i])
		}
	}

	return status, str.String()
}
