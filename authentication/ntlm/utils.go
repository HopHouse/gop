package ntlm

import (
	"bytes"
	"fmt"
	"strings"
	"syscall"
)

// Taken from "golang.org/x/sys/windows"
// ByteSliceFromString returns a NUL-terminated slice of bytes
// containing the text of s. If s contains a NUL byte at any
// location, it returns (nil, syscall.EINVAL).
func ByteSliceFromString(s string) ([]byte, error) {
	if strings.IndexByte(s, 0) != -1 {
		return nil, syscall.EINVAL
	}
	a := make([]byte, len(s)+1)
	copy(a, s)
	return a, nil
}

// Taken from "golang.org/x/sys/windows"
// ByteSliceToString returns a string form of the text represented by the slice s, with a terminating NUL and any
// bytes after the NUL removed.
func ByteSliceToString(s []byte) string {
	if i := bytes.IndexByte(s, 0); i != -1 {
		s = s[:i]
	}
	return string(s)
}

// Return :
// - 0 if equals
// - 1 if len(s1) > len(s2)
// - -1 if len(s1) < len(s2)
// - 2 if bitwise comparaison of s1 != s2
// TODO Mutualise with the one in the relay package
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
