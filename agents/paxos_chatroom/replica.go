package paxos

// This file contains the definition and logic of a paxos agent.
// The ReplicaAgent type must implement the Agent interface.

import (
	"fmt"
	"strconv"
	"strings"
	"time"

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

	myID c.ProcessID
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
func (rep *ReplicaAgent) Init(attrs map[string]interface{},
	send func(vDest c.ProcessID, msg string),
	fatalAgentErrorf func(errMsg string, a ...interface{}),
	debugPrintf func(s string, a ...interface{})) {
	rep.send = send
	rep.fatalAgentErrorf = fatalAgentErrorf
	rep.debugPrintf = debugPrintf
	rep.isActive = false
	rep.inMemStore = make(map[string]string)

	rep.myID = c.ProcessID(attrs["myid"].(float64))
}

// Halt stops the execution of paxos.
func (rep *ReplicaAgent) Halt() {
	rep.isActive = false
}

// Deliver a message
func (rep *ReplicaAgent) Deliver(request string, port c.PortNum) {
	switch port {
	case 2:
		// Command from client, format "<clientID> <reqNum> <m>"
		reqSlice := strings.SplitN(request, " ", 3)
		id, err := strconv.ParseUint(reqSlice[0], 10, 64)
		if err != nil {
			rep.fatalAgentErrorf("Invalid request %s\n", request)
		}

		// TODO: Now just simulate processing and send reply
		clientID := c.ProcessID(id)
		reqNum, _ := strconv.ParseUint(reqSlice[1], 10, 64)
		m := reqSlice[2]
		time.Sleep(timeoutDuration / 2)
		rep.debugPrintf("Replica %d commited (%d, %d, %s)\n",
			rep.myID, clientID, reqNum, m)
		// Send response  "committed <clientID> <reqNum>"
		rep.send(clientID, fmt.Sprintf("committed %d %d", clientID, reqNum))
	case 9:
		// Command from controller
		rep.debugPrintf("%s\n", request)
	default:
		rep.fatalAgentErrorf("Received '%s' in unexpected port %v\n", request, port)
	}
}

// Run begins the execution of the paxos agent.
func (rep *ReplicaAgent) Run() {
	rep.isActive = true
}
