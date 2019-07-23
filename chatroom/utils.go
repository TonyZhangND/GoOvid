package main

import (
	"fmt"
	"os"
	"runtime/debug"
	"sync"
	"time"
)

const pingInterval = 500 * time.Millisecond
const basePort = 3000

type processID uint16

// *****  MESSAGE LOG *****

type messageLog struct {
	log []string
	sync.RWMutex
}

func newMessageLog() *messageLog {
	return &messageLog{log: make([]string, 0)}
}

// Appends message m to the log of ml
func (ml *messageLog) appendMsg(m string) {
	ml.Lock()
	ml.log = append(ml.log, m)
	ml.Unlock()
}

// Returns a copy of the log of ml
func (ml *messageLog) getMessages() []string {
	result := make([]string, len(ml.log))
	ml.RLock()
	for i, m := range ml.log {
		result[i] = m
	}
	ml.RUnlock()
	return result
}

// *****  UTILITIES *****

// Prints the string s if debug mode is on
func debugPrintln(s string) {
	if debugMode {
		fmt.Printf("Process %v : %v\n", myPhysID, s)
	}
}

func fatalError(errMsg string) {
	shouldRun = false
	fmt.Printf("Error : process %v : %v\n", myPhysID, errMsg)
	debug.PrintStack()
	os.Exit(1)
}
