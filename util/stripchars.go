package util

import "strings"

// Stripchars removes the characters from chr in str and returns the updated string.
func Stripchars(str, chr string) string {
	return strings.Map(func(r rune) rune {
		if !strings.ContainsRune(chr, r) {
			return r
		}
		return -1
	}, str)
}
