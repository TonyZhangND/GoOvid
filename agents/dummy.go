package agents

import (
	"fmt"
	"strings"

	c "github.com/TonyZhangND/GoOvid/commons"
)

// DummyAgent struct
type DummyAgent struct {
	send             func(vDest c.ProcessID, msg string)
	fatalAgentErrorf func(errMsg string, a ...interface{})
	debugPrintf      func(s string, a ...interface{})
	isActive         bool
}

// Init fills the empty struct with this agent's fields and attributes.
func (da *DummyAgent) Init(attrs map[string]interface{},
	send func(vDest c.ProcessID, msg string),
	fatalAgentErrorf func(errMsg string, a ...interface{}),
	debugPrintf func(s string, a ...interface{})) {
	da.send = send
	da.fatalAgentErrorf = fatalAgentErrorf
}

// Halt stops the execution of *da.
func (da *DummyAgent) Halt() {
}

// Deliver a message of the format "<sender name> <contents>".
// The chat agent ignores the port.
func (da *DummyAgent) Deliver(data string, port c.PortNum) {
	dataSlice := strings.SplitN(strings.TrimSpace(data), " ", 2)
	sender, msg := dataSlice[0], dataSlice[1]
	fmt.Printf("%s > %s\n", sender, msg)
}

// Run begins the execution of the *da agent.
func (da *DummyAgent) Run() {}
