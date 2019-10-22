package paxos

// This file contains the definition and logic of a client using the paxos service.
// The ClientAgent type must implement the Agent interface.

import (
	c "github.com/TonyZhangND/GoOvid/commons"
)

// ClientAgent struct contains the information inherent to a paxos client
type ClientAgent struct {
	send             func(vDest c.ProcessID, msg string)
	fatalAgentErrorf func(errMsg string, a ...interface{})
	debugPrintf      func(s string, a ...interface{})
	isActive         bool
}

// Init fills the empty client struct with this agent's fields and attributes.
func (client *ClientAgent) Init(attrs map[string]interface{},
	send func(vDest c.ProcessID, msg string),
	fatalAgentErrorf func(errMsg string, a ...interface{}),
	debugPrintf func(s string, a ...interface{})) {
	client.send = send
	client.fatalAgentErrorf = fatalAgentErrorf
	client.debugPrintf = debugPrintf
	client.isActive = false
}

// Halt stops the execution of the agent.
func (client *ClientAgent) Halt() {
	client.isActive = false
}

// Deliver a message
func (client *ClientAgent) Deliver(request string, port c.PortNum) {
	switch port {
	case 9:
		// Command from controller
		client.debugPrintf("%s\n", request)
	default:
		client.fatalAgentErrorf("Received '%s' in unexpected port %v", request, port)
	}
}

// Run begins the execution of the paxos agent.
func (client *ClientAgent) Run() {
	client.isActive = true
}
