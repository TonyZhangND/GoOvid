package paxos

// This file contains the definition and logic of a paxos agent.
// The ReplicaAgent type must implement the Agent interface.

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	c "github.com/TonyZhangND/GoOvid/commons"
)

// ReplicaAgent struct contains the information inherent to a paxos replica
type ReplicaAgent struct {
	// Default GoOvid states
	send             func(vDest c.ProcessID, msg string)
	fatalAgentErrorf func(errMsg string, a ...interface{})
	debugPrintf      func(s string, a ...interface{})
	isActive         bool

	// Replica attributes
	myID     c.ProcessID
	replicas map[c.ProcessID]int
	clients  map[c.ProcessID]int
	mode     string // script or manual modes
	output   string // path to output file for 'dump' command
	log      string // TODO: This doesn't do anything now

	// Replica state
	chatLog   []string // application state
	slotIn    uint64
	slotOut   uint64
	requests  map[string]*request  // given k->*v, k is a hash of v
	proposals map[string]*proposal // given k->*v, k is a hash of v
	decisions map[uint64]*request  // map of slot -> decision
	// failureDetector *unreliableFailureDetector // TODO: currently not in use
	// acceptor *acceptorState
	// leader   *leaderState
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

	// Initialize replica attributes
	rep.myID = c.ProcessID(attrs["myid"].(float64))
	rep.replicas = make(map[c.ProcessID]int)
	for _, x := range attrs["replicas"].([]interface{}) {
		id := c.ProcessID(x.(float64))
		rep.replicas[id] = 0
	}
	rep.clients = make(map[c.ProcessID]int)
	for _, x := range attrs["clients"].([]interface{}) {
		id := c.ProcessID(x.(float64))
		rep.clients[id] = 0
	}
	rep.log = attrs["log"].(string)
	rep.output = attrs["output"].(string)

	// Initialize replica state
	rep.chatLog = make([]string, 0)
	rep.slotIn = 0
	rep.slotOut = 0
	rep.requests = make(map[string]*request)
	rep.proposals = make(map[string]*proposal)
	rep.decisions = make(map[uint64]*request)
}

// Halt stops the execution of paxos.
func (rep *ReplicaAgent) Halt() {
	rep.isActive = false
}

// Deliver a message
func (rep *ReplicaAgent) Deliver(request string, port c.PortNum) {
	switch port {
	case 1:
		// Message from another replica
		msgHeader := strings.SplitN(request, " ", 2)[0]
		switch msgHeader {
		case "decision":
			rep.handleClientRequest(request)
		default:
			rep.fatalAgentErrorf("Received invalid msg '%s'\n", request)
		}

	case 2:
		// Command from client, format "<clientID> <reqNum> <m>"
		rep.handleClientRequest(request)
	case 9:
		// Command from controller
		rep.handleControllerCommand(request)
	default:
		rep.fatalAgentErrorf("Received '%s' in unexpected port %v\n", request, port)
	}
}

// Run begins the execution of the paxos agent.
func (rep *ReplicaAgent) Run() {
	rep.isActive = true
}

func (rep *ReplicaAgent) handleControllerCommand(r string) {
	cmd := strings.SplitN(r, " ", 2)[0]
	switch cmd {
	case "dump":
		rep.debugPrintf("Handle dump\n")
		f, err := os.Create(rep.output)
		defer f.Close()
		if err != nil {
			rep.fatalAgentErrorf("Error creating file %s: %v\n", rep.output, err)
		}
		w := bufio.NewWriter(f)
		for _, s := range rep.chatLog {
			_, err = w.WriteString(fmt.Sprintf("%s\n", s))
			if err != nil {
				rep.fatalAgentErrorf("Error writing to file %s: %v\n", rep.output, err)
			}
		}
		w.Flush()
	case "kill":
		// TODO
	case "skip":
		// TODO
	}
}

// Handles an incoming client request "<clientID> <reqNum> <m>"
func (rep *ReplicaAgent) handleClientRequest(r string) {
	reqSlice := strings.SplitN(r, " ", 3)
	cid, _ := strconv.ParseUint(reqSlice[0], 10, 64)
	rn, _ := strconv.ParseUint(reqSlice[1], 10, 64)
	m := reqSlice[2]
	req := &request{c.ProcessID(cid), rn, m}

	// Add req to my handy dandy set of requests and propose()
	rep.requests[req.hash()] = req
	rep.propose()
}

// Propose method in Fig 1 of PMMC
func (rep *ReplicaAgent) propose() {
	for k, req := range rep.requests {
		// For each req in rep.requests, start proposing it for each slot
		// that I have not proposed a value nor learned a decision
		_, slotTaken := rep.decisions[rep.slotIn]
		for slotTaken {
			rep.slotIn++
			_, slotTaken = rep.decisions[rep.slotIn]
		}
		// Found an empty slot
		delete(rep.requests, k)
		prop := &proposal{rep.slotIn, req}
		rep.proposals[prop.hash()] = prop
		rep.slotIn++
	}
}

// Perform method in Fig 1 of PMMC
func (rep *ReplicaAgent) perform(req *request) {
	for s, dec := range rep.decisions {
		// If req has been previously committed, ignore it
		if s < rep.slotOut && dec.hash() == req.hash() {
			rep.slotOut++
			return
		}
	}
	// Else execute the request
	rep.chatLog = append(rep.chatLog, fmt.Sprintf("%d: %s", req.clientID, req.payload))
}

// Handles a decision message "decision <slot> <clientID> <reqNum> <m>"
func (rep *ReplicaAgent) handleDecision(d string) {
	// Store decision in rep.decisions
	dSlice := strings.SplitN(d, " ", 5)
	slot, _ := strconv.ParseUint(dSlice[1], 10, 64)
	id, _ := strconv.ParseUint(dSlice[2], 10, 64)
	cid := c.ProcessID(id)
	reqNum, _ := strconv.ParseUint(dSlice[3], 10, 64)
	m := dSlice[4]
	newDec := &request{cid, reqNum, m}
	rep.decisions[slot] = newDec

	// Execute all decisions that can be committed
	decToExec, ok := rep.decisions[rep.slotOut]
	for ok {
		// If slot of request I am about to excute is used in proposals, then
		// 1. remove it from proposals, and
		// 2. if the req removed is not the one I am about to execute, put it back
		//    into rep.requests
		for k, prop := range rep.proposals {
			if prop.slot == rep.slotOut {
				// If slotOut used for a command in rep.proposals
				delete(rep.proposals, k)
				if prop.hash() != decToExec.hash() {
					// If req removed from rep.proposals is not decToExec
					rep.requests[prop.hash()] = prop.req
				}
				break // No need to keep searching
			}
		}
		rep.perform(decToExec)
		decToExec, ok = rep.decisions[rep.slotOut]
	}
	rep.propose()
}
