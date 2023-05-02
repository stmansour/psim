package util

import "fmt"

// DPrintf is the debug logger
func DPrintf(format string, a ...interface{}) {
	fmt.Printf("**DBG** ")
	fmt.Printf(format, a...)
}
