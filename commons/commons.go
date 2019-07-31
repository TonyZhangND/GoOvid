package commons

import (
	"fmt"
	"os"
	"runtime/debug"
)

type (
	// ProcessID is a type representing the physical and virtual ID's of an agent
	ProcessID uint16
	// PortNum is a type representing an IP port on a host
	PortNum uint16
)

// FatalOvidErrorf prints the error messange and kills the entire program
func FatalOvidErrorf(s string, a ...interface{}) {
	errMsg := fmt.Sprintf(s, a...)
	fmt.Printf("Error : Ovid : %v", errMsg)
	debug.PrintStack()
	os.Exit(1)
}

// CheckFatalOvidErrorf prints the error messange and kills the entire program
// if an error is detected
func CheckFatalOvidErrorf(err error, s string, a ...interface{}) {
	if err != nil {
		FatalOvidErrorf(s, a...)
		os.Exit(1)
	}
}
