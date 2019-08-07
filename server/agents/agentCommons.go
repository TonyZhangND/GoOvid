package agents

import (
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
	// Init populates an empty struct for the agent
	Init(attrs map[string]interface{},
		send func(dest c.ProcessID, msg string),
		fatalAgentErrorf func(errMsg string, a ...interface{}))
	// Run starts the agent's main loop, if any
	Run()
	// Deliver delivers msg to the agent
	Deliver(msg string)
	// Stops the agent from processing new messages
	Halt()
}

// AgentInfo is a struct containing data common to all agents
type AgentInfo struct {
	Type     AgentType
	Box      c.BoxID
	RawAttrs map[string]interface{}
	Routes   map[c.ProcessID]c.Route
}
