package agents

import (
	"fmt"
	"os"
	"runtime/debug"

	c "github.com/TonyZhangND/GoOvid/commons"
)

const debugMode = true

// AgentType is an integer denoting the type of an agent
// It acts like an enum
type AgentType int

const (
	// Chat agent enum
	Chat AgentType = iota
)

// Agent is an interface that all agents must implement
type Agent interface {
	Init(attrs map[string]interface{}, broadcast func(msg string))
	Run()
	Deliver(msg string)
	Halt()
	Name() string
}

// AgentInfo is a struct containing data common to all agents
type AgentInfo struct {
	Type     AgentType
	Box      c.BoxID
	RawAttrs map[string]interface{}
	Routes   map[c.ProcessID]c.Route
}

// Prints the string s if debug mode is on
func debugPrintln(agent *Agent, s string) {
	if debugMode {
		fmt.Printf("Agent %v : %v\n", (*agent).Name(), s)
	}
}

// Prints the error messange and halts the agent
func fatalAgentErrorf(agent Agent, errMsg string, a ...interface{}) {
	agent.Halt()
	msg := fmt.Sprintf(errMsg, a...)
	fmt.Printf("Error : Agent %v : %v", agent.Name(), msg)
	debug.PrintStack()
	os.Exit(1)
}
