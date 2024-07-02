package util

import (
	"fmt"
	"unicode"
)

// PrintHexAndASCII prints a buffer in hex and ASCII format
func PrintHexAndASCII(buffer []byte, maxChars int) {
	if maxChars < 1 || maxChars > len(buffer) {
		maxChars = len(buffer)
	}

	for i := 0; i < maxChars; i += 16 {
		// Print hex values
		for j := 0; j < 16; j++ {
			if i+j < maxChars {
				fmt.Printf("%02X ", buffer[i+j])
			} else {
				fmt.Print("   ")
			}
		}

		// Print ASCII values
		fmt.Print(" | ")
		for j := 0; j < 16; j++ {
			if i+j < maxChars {
				b := buffer[i+j]
				if unicode.IsPrint(rune(b)) {
					fmt.Printf("%c", b)
				} else {
					fmt.Print(".")
				}
			}
		}
		fmt.Println()
	}
}
