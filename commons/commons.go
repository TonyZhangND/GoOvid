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

// FatalOvidError prints the error messange and kills the entire program
func FatalOvidError(errMsg string) {
	fmt.Printf("Error : Ovid : %v\n", errMsg)
	debug.PrintStack()
	os.Exit(1)
}

// CheckFatalOvidError prints the error messange and kills the entire program
// if an error is detected
func CheckFatalOvidError(err error, errMsg string) {
	if err != nil {
		fmt.Printf("Error : Ovid : %v\n", errMsg)
		debug.PrintStack()
		os.Exit(1)
	}
}
