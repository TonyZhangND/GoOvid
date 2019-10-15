package agents

// This file contains the definition and logic of a paxos agent.
// The PaxosAgent type must implement the Agent interface.
// Requirement: keys do not contain whitespace

import (
	mapset "github.com/deckarep/golang-set"

	c "github.com/TonyZhangND/GoOvid/commons"
)

// PaxosAgent struct contains the information inherent to a paxos agent
type PaxosAgent struct {
	send             func(vDest c.ProcessID, msg string)
	fatalAgentErrorf func(errMsg string, a ...interface{})
	debugPrintf      func(s string, a ...interface{})
	isActive         bool
	inMemStore       map[string]string
	replica          replicaState
	acceptor         acceptorState
	leader           leaderState
}

type replicaState struct {
	inMemStore map[string]string
	slotIn     uint
	slotOut    uint
	requests   mapset.Set
	proposals  mapset.Set
	decisions  mapset.Set
}

type acceptorState struct {
	ballotNum ballot
	accepted  mapset.Set
}

type leaderState struct {
	ballotNum ballot
	accepted  mapset.Set
}

type ballot struct {
	num       uint
	leaderID  c.ProcessID
	active    bool
	proposals mapset.Set
}

// Init fills the empty kvs struct with this agent's fields and attributes.
func (paxos *PaxosAgent) Init(attrs map[string]interface{},
	send func(vDest c.ProcessID, msg string),
	fatalAgentErrorf func(errMsg string, a ...interface{}),
	debugPrintf func(s string, a ...interface{})) {
	paxos.send = send
	paxos.fatalAgentErrorf = fatalAgentErrorf
	paxos.debugPrintf = debugPrintf
	paxos.isActive = false
	paxos.inMemStore = make(map[string]string)

}

// Halt stops the execution of paxos.
func (paxos *PaxosAgent) Halt() {
	paxos.isActive = false
}

// Deliver a message
func (paxos *PaxosAgent) Deliver(request string, port c.PortNum) {

}

// Run begins the execution of the paxos agent.
func (paxos *PaxosAgent) Run() {
	paxos.isActive = true
}
