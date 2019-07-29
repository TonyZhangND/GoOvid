package server

// This file contains some utility procedures

import (
	"fmt"
	"os"
	"runtime/debug"
	"time"
)

const debugMode = true
const pingInterval = 500 * time.Millisecond
const basePort = 3000

type processID uint16

// DebugPrintln prints the string s if debug mode is on
func debugPrintln(s string) {
	if debugMode {
		fmt.Printf("Process %v : %v\n", myPhysID, s)
	}
}

// fatalServerError prints the error messange and kills the program
func fatalServerError(errMsg string) {
	shouldRun = false
	fmt.Printf("Error : process %v : %v\n", myPhysID, errMsg)
	debug.PrintStack()
	os.Exit(1)
}

// checkFatalServerError prints the error messange and kills the program
// if an error is detected
func checkFatalServerError(e error, errMsg string) {
	if e != nil {
		fatalServerError(errMsg)
	}
}
