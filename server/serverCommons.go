package server

// This file contains some utility procedures

import (
	"fmt"
	"os"
	"runtime/debug"
	"time"

	c "github.com/TonyZhangND/GoOvid/commons"
)

const pingInterval = 500 * time.Millisecond
const basePort c.PortNum = 3000

// DebugMode turns on debugging print statements when true
var DebugMode = false

// DebugPrintln prints the string s if debug mode is on
// agentID is -1 when agent id is irrelevant in the context at which
// this function is called.
func debugPrintf(agentID int, s string, a ...interface{}) {
	if DebugMode {
		errMsg := fmt.Sprintf(s, a...)
		if agentID < 0 {
			fmt.Printf("Box %v : %s", myBoxID, errMsg)
		} else {
			fmt.Printf("Box %v, Agent %d : %s", myBoxID, agentID, errMsg)
		}
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
