package agents

import (
	"fmt"
	"os"
	"runtime/debug"
)

const debugMode = true

type AgentType int

const (
	Chat AgentType = 1
)

type Agent interface {
	deliver(msg string)
	run()
	halt()
	name() string
}

// DebugPrintln prints the string s if debug mode is on
func debugPrintln(agent *Agent, s string) {
	if debugMode {
		fmt.Printf("Agent %v : %v\n", (*agent).name())
	}
}

// fatalServerError prints the error messange and halts the agent
func fatalError(agent Agent, errMsg string) {
	agent.halt()
	fmt.Printf("Error : Agent %v : %v\n", agent.name(), errMsg)
	debug.PrintStack()
	os.Exit(1)
}
