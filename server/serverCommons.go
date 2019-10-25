package server

// This file contains some utility procedures

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"time"

	c "github.com/TonyZhangND/GoOvid/commons"
)

const pingInterval = 500 * time.Millisecond
const basePort c.PortNum = 3000

// DebugMode turns on debugging print statements when true
var DebugMode = false

// LogFile turns on logging when initialized
var LogFile = ""

// DebugPrintln prints the string s if debug mode is on
func debugPrintf(s string, a ...interface{}) {
	if DebugMode {
		errMsg := fmt.Sprintf(s, a...)
		fmt.Printf("Box %v : %s", myBoxID, errMsg)
	}
	if len(LogFile) > 0 {
		f, err := os.OpenFile(LogFile,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Println(err)
		}
		defer f.Close()
		errMsg := fmt.Sprintf(s, a...)
		s := fmt.Sprintf("Box %v : %s", myBoxID, errMsg)
		if _, err := f.WriteString(s); err != nil {
			log.Println(err)
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
