package server

import (
	"fmt"
	"os"
	"runtime/debug"
	"time"
)

const pingInterval = 700 * time.Millisecond
const basePort = 3000

type processID uint16

// *****  UTILITIES *****

// Prints the string s if debug mode is on
func debugPrintln(s string) {
	if debugMode {
		fmt.Printf("%v\n", s)
	}
}

func fatalError(errMsg string) {
	shouldRun = false
	fmt.Printf("Error : process %v : %v\n", myPhysID, errMsg)
	debug.PrintStack()
	os.Exit(1)
}
