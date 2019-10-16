package paxos

// This file contains the definition and logic of a centralized controller of the
// paxos service.
// The ControllerAgent type must implement the Agent interface.

import (
	c "github.com/TonyZhangND/GoOvid/commons"
)

// ControllerAgent struct contains the information inherent to a controller
type ControllerAgent struct {
	send             func(vDest c.ProcessID, msg string)
	fatalAgentErrorf func(errMsg string, a ...interface{})
	debugPrintf      func(s string, a ...interface{})
	isActive         bool
}

// Init fills the empty ctr struct with this agent's fields and attributes.
func (ctr *ControllerAgent) Init(attrs map[string]interface{},
	send func(vDest c.ProcessID, msg string),
	fatalAgentErrorf func(errMsg string, a ...interface{}),
	debugPrintf func(s string, a ...interface{})) {
	ctr.send = send
	ctr.fatalAgentErrorf = fatalAgentErrorf
	ctr.debugPrintf = debugPrintf
	ctr.isActive = false
}

// Halt stops the execution of the agent.
func (ctr *ControllerAgent) Halt() {
	ctr.isActive = false
}

// Deliver a message
func (ctr *ControllerAgent) Deliver(request string, port c.PortNum) {

}

// Run begins the execution of the paxos agent.
func (ctr *ControllerAgent) Run() {
	ctr.isActive = true
}
