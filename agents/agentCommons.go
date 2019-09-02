package agents

// This file contains the definitions common to any agent.
// In particular, it contains the Agent interface that all agents must implement.

import (
	c "github.com/TonyZhangND/GoOvid/commons"
)

// AgentType is an integer denoting the type of an agent
// It acts like an enum
type AgentType int

const (
	// Dummy agent enum
	Dummy AgentType = iota
	// Chat agent enum
	Chat AgentType = iota
	// KVS agent enum
	KVS AgentType = iota
	// Client agent enum
	Client AgentType = iota
	// TTY agent enum
	TTY AgentType = iota
)

// Agent is an interface that all agents must implement
type Agent interface {
	// Init populates an empty struct for the agent
	// - attrs is a map containing the attributes of the agent
	// - send is a function that the agent calls to send msg to virtual receiver vDest
	// - fatalAgentErrorf is a function that halts the agent's operation and prints
	//   the error stack.
	// - debugPrintf is a function that prints some debugging message if debugMode is on
	Init(attrs map[string]interface{},
		send func(vDest c.ProcessID, msg string),
		fatalAgentErrorf func(errMsg string, a ...interface{}),
		debugPrintf func(s string, a ...interface{}))

	// Run starts the agent's main loop, if any
	Run()

	// Deliver delivers msg to the agent at the specified port
	Deliver(data string, port c.PortNum)

	// Stops the agent from processing new messages
	Halt()
}

// AgentInfo is a struct containing data common to all agents.
// It corresponds to the format of a JSON entry for an agent configuration.
type AgentInfo struct {
	Type     AgentType
	Box      c.BoxID
	RawAttrs map[string]interface{}
	Routes   map[c.ProcessID]c.Route
}

// NewAgent returns a new, empty struct corresponding to the agent type t
func NewAgent(t AgentType) Agent {
	switch t {
	case Chat:
		return &ChatAgent{}
	case Dummy:
		return &DummyAgent{}
	case KVS:
		return &KVSAgent{}
	case Client:
		return &ClientAgent{}
	case TTY:
		return &TTYAgent{}
	default:
		c.FatalOvidErrorf("Invalid agent type for agent %v\n", t)
		return nil
	}
}
