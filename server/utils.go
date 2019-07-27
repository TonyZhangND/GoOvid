package server

// This file contains the definitions of global constants, as well as
// some utility procedures

import (
	"fmt"
	"os"
	"runtime/debug"
	"time"
)

const pingInterval = 500 * time.Millisecond
const basePort = 3000

type processID uint16

// Prints the string s if debug mode is on
func debugPrintln(s string) {
	if debugMode {
		fmt.Printf("Process %v : %v\n", myPhysID, s)
	}
}

// Prints the error messange and kills the program
func fatalError(errMsg string) {
	shouldRun = false
	fmt.Printf("Error : process %v : %v\n", myPhysID, errMsg)
	debug.PrintStack()
	os.Exit(1)
}
