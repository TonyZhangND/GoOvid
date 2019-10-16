package paxos

// This file contains the definition and logic of a paxos agent.
// The ReplicaAgent type must implement the Agent interface.

import (
	mapset "github.com/deckarep/golang-set"

	c "github.com/TonyZhangND/GoOvid/commons"
)

// ReplicaAgent struct contains the information inherent to a paxos replica
type ReplicaAgent struct {
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
func (replica *ReplicaAgent) Init(attrs map[string]interface{},
	send func(vDest c.ProcessID, msg string),
	fatalAgentErrorf func(errMsg string, a ...interface{}),
	debugPrintf func(s string, a ...interface{})) {
	replica.send = send
	replica.fatalAgentErrorf = fatalAgentErrorf
	replica.debugPrintf = debugPrintf
	replica.isActive = false
	replica.inMemStore = make(map[string]string)

}

// Halt stops the execution of paxos.
func (replica *ReplicaAgent) Halt() {
	replica.isActive = false
}

// Deliver a message
func (replica *ReplicaAgent) Deliver(request string, port c.PortNum) {

}

// Run begins the execution of the paxos agent.
func (replica *ReplicaAgent) Run() {
	replica.isActive = true
}
