package server

// This file contains some utility procedures

import (
	"fmt"
	"os"
	"runtime/debug"
	"time"

	c "github.com/TonyZhangND/GoOvid/commons"
)

const debugMode = true
const pingInterval = 500 * time.Millisecond
const basePort c.PortNum = 3000

// DebugPrintln prints the string s if debug mode is on
func debugPrintf(s string, a ...interface{}) {
	if debugMode {
		errMsg := fmt.Sprintf(s, a...)
		fmt.Printf("Box %v : %s", myBoxID, errMsg)
	}
}

// fatalServerErrorf prints the error messange and kills the program
func fatalServerErrorf(s string, a ...interface{}) {
	shouldRun = false
	msg := fmt.Sprintf(s, a...)
	fmt.Printf("Error : process %v : %v", myBoxID, msg)
	debug.PrintStack()
	os.Exit(1)
}

// Prints the error messange and kills the program
// if an error is detected
func checkFatalServerErrorf(e error, s string, a ...interface{}) {
	if e != nil {
		fatalServerErrorf(s, a...)
	}
}
