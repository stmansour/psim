package util

import "fmt"

// DPrintf is the debug logger
func DPrintf(format string, a ...interface{}) {
	fmt.Printf("*** DEBUG *** ")
	fmt.Printf(format, a...)
}
