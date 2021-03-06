package paxos

// This file contains the definition and logic of a paxos agent.
// The ReplicaAgent type must implement the Agent interface.

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"

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
	myID      c.ProcessID
	replicas  map[c.ProcessID]int
	clients   map[c.ProcessID]int
	mode      string         // script or manual modes
	output    string         // path to output file for 'dump' command
	skipSlots map[uint64]int // set containing slots to skip, see spec

	// Replica state
	chatLog   []string // application state
	slotIn    uint64
	slotOut   uint64
	requests  map[string]*request  // given k->*v, k is a hash of v
	proposals map[string]*proposal // given k->*v, k is a hash of v
	decisions map[uint64]*request  // map of slot -> decision

	rmut *sync.RWMutex // mutex for requests map
	pmut *sync.RWMutex // mutex for proposals map
	dmut *sync.RWMutex //mutex for decisions map

	failureDetector *unreliableFailureDetector // TODO: currently only used to mark leaders
	acceptor        *acceptorState
	leader          *leaderState
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
	rep.output = attrs["output"].(string)
	rep.skipSlots = make(map[uint64]int)
	if _, ok := attrs["skip"].([]interface{}); ok {
		for _, x := range attrs["skip"].([]interface{}) {
			slot := uint64(x.(float64))
			rep.skipSlots[slot] = 0
		}
	}
	rep.debugPrintf("Skipping these slots : %v\n", rep.skipSlots)

	// Initialize replica state
	rep.chatLog = make([]string, 0)
	rep.slotIn = 0
	rep.slotOut = 0
	rep.requests = make(map[string]*request)
	rep.proposals = make(map[string]*proposal)
	rep.decisions = make(map[uint64]*request)
	rep.rmut = new(sync.RWMutex) // mutex for requests map
	rep.pmut = new(sync.RWMutex) // mutex for requests map
	rep.dmut = new(sync.RWMutex) // mutex for requests map
	rep.acceptor = rep.newAcceptorState()
	rep.leader = rep.newLeaderState()
	rep.failureDetector = newUnreliableFailureDetector(rep)
	for id := range rep.replicas {
		// TODO: Just make everyone leaders for now
		rep.failureDetector.leaders[id] = true
	}
}

// Halt stops the execution of paxos.
func (rep *ReplicaAgent) Halt() {
	rep.isActive = false
}

// Run begins the execution of the paxos agent.
func (rep *ReplicaAgent) Run() {
	rep.isActive = true
	rep.runLeader()
}

// Deliver a message
func (rep *ReplicaAgent) Deliver(request string, port c.PortNum) {
	switch port {
	case 1:
		// Message from another replica
		msgHeader := strings.SplitN(request, " ", 2)[0]
		switch msgHeader {
		case "decision":
			rep.handleDecision(request)
		case "p1a":
			rep.handleP1a(request)
		case "p2a":
			rep.handleP2a(request)
		case "p1b":
			rep.handleP1b(request)
		case "p2b":
			rep.handleP2b(request)
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

func (rep *ReplicaAgent) dumpPaxosLog() {
	rep.debugPrintf("Handle dump\n")
	f, err := os.Create(rep.output)
	defer f.Close()
	if err != nil {
		rep.fatalAgentErrorf("Error creating file %s: %v\n", rep.output, err)
	}
	w := bufio.NewWriter(f)
	if len(rep.decisions) == 0 {
		return
	}
	keys := make([]int, len(rep.decisions))
	i := 0
	for k := range rep.decisions {
		keys[i] = int(k)
		i++
	}
	sort.Ints(keys)
	maxSlotFilled := keys[len(keys)-1]
	for i = 0; i <= maxSlotFilled; i++ {
		if prop, ok := rep.decisions[uint64(i)]; ok {
			_, err := w.WriteString(fmt.Sprintf("%d, %d, '%s'\n", prop.clientID, prop.reqNum, prop.payload))
			if err != nil {
				rep.fatalAgentErrorf("Error writing to file %s: %v\n", rep.output, err)
			}
		} else {
			_, err := w.WriteString("hole\n")
			if err != nil {
				rep.fatalAgentErrorf("Error writing to file %s: %v\n", rep.output, err)
			}
		}
	}
	w.Flush()
}

func (rep *ReplicaAgent) dumpChatLog() {
	rep.debugPrintf("Handle dump the stupid way\n")
	f, err := os.Create(rep.output)
	defer f.Close()
	if err != nil {
		rep.fatalAgentErrorf("Error creating file %s: %v\n", rep.output, err)
	}
	if len(rep.decisions) == 0 {
		return
	}
	keys := make([]int, len(rep.decisions))
	i := 0
	for k := range rep.decisions {
		keys[i] = int(k)
		i++
	}
	sort.Ints(keys)
	firstHole := len(keys) - 1
	for i := 0; i < len(keys); i++ {
		if keys[i] != i {
			firstHole = i
			break
		}
	}
	rep.debugPrintf("First hole is %d\n", firstHole)

	seen := make(map[string]bool)
	chatLog := make([]string, 0)
	rep.dmut.RLock()
	for i := 0; i < firstHole; i++ {
		s := uint64(keys[i])
		if _, ok := seen[rep.decisions[s].hash()]; !ok {
			log := fmt.Sprintf("%d, %d : '%s'\n", rep.decisions[s].clientID, rep.decisions[s].reqNum, rep.decisions[s].payload)
			chatLog = append(chatLog, log)
		}
		seen[rep.decisions[s].hash()] = true
	}
	rep.dmut.RUnlock()
	w := bufio.NewWriter(f)
	for _, c := range chatLog {
		_, err := w.WriteString(c)
		if err != nil {
			rep.fatalAgentErrorf("Error writing to file %s: %v\n", rep.output, err)
		}
	}
	w.Flush()
}

func (rep *ReplicaAgent) handleControllerCommand(r string) {
	cmd := strings.SplitN(r, " ", 2)[0]
	switch cmd {
	case "dump":
		rep.dumpChatLog()
		// rep.debugPrintf("Handle dump\n")
		// // rep.debugPrintf("LOG %v\n", rep.chatLog)
		// f, err := os.Create(rep.output)
		// // rep.debugPrintf("HELLO1\n")
		// defer f.Close()
		// if err != nil {
		// 	rep.fatalAgentErrorf("Error creating file %s: %v\n", rep.output, err)
		// }
		// w := bufio.NewWriter(f)
		// // rep.debugPrintf("HELLO2 %d\n", len(rep.chatLog))
		// for _, s := range rep.chatLog {
		// 	rep.debugPrintf("%s\n", s)
		// 	_, err = w.WriteString(fmt.Sprintf("%s\n", s))
		// 	if err != nil {
		// 		rep.fatalAgentErrorf("Error writing to file %s: %v\n", rep.output, err)
		// 	}
		// }
		// // rep.debugPrintf("BOB\n")
		// w.Flush()

	case "kill":
		// TODO
	case "skip":
		// TODO
	}
}

// Handles an incoming client request "<clientID> <reqNum> <m>"
func (rep *ReplicaAgent) handleClientRequest(r string) {
	if _, ok := rep.failureDetector.leaders[rep.myID]; !ok {
		// ignore request if I am not leader
		return
	}
	reqSlice := strings.SplitN(r, " ", 3)
	cid, _ := strconv.ParseUint(reqSlice[0], 10, 64)
	rn, _ := strconv.ParseUint(reqSlice[1], 10, 64)
	m := reqSlice[2]
	req := &request{c.ProcessID(cid), rn, m}

	// If request is already decided, return the decision
	rep.dmut.RLock()
	for _, decision := range rep.decisions {
		if decision.eq(req) {
			defer rep.dmut.RUnlock()
			response := fmt.Sprintf("committed %d %d", req.clientID, req.reqNum)
			rep.send(req.clientID, response)
			return
		}
	}
	rep.dmut.RUnlock()

	// Add req to my handy dandy set of requests only if it is not repeated, and propose()
	isOldReq := false
	rep.rmut.RLock()
	for _, myReq := range rep.requests {
		if req.eq(myReq) {
			isOldReq = true
		}
	}
	rep.rmut.RUnlock()
	rep.pmut.RLock()
	if !isOldReq {
		for _, p := range rep.proposals {
			if req.eq(p.req) {
				isOldReq = true
			}
		}
	}
	rep.pmut.RUnlock()
	if !isOldReq {
		rep.rmut.Lock()
		rep.requests[req.hash()] = req
		rep.rmut.Unlock()
	}
	rep.propose()
}

// Propose method in Fig 1 of PMMC
func (rep *ReplicaAgent) propose() {
	rep.rmut.Lock()
	for k, req := range rep.requests {
		// For each req in rep.requests, start proposing it for each slot
		// that I have not proposed a value nor learned a decision
		rep.dmut.RLock()
		_, slotTaken := rep.decisions[rep.slotIn]
		rep.dmut.RUnlock()
		s, skipSlot := rep.skipSlots[rep.slotIn]
		for slotTaken || skipSlot {
			if skipSlot {
				rep.debugPrintf("Skipping slot %d\n", s)
			}
			rep.slotIn++
			rep.dmut.RLock()
			_, slotTaken = rep.decisions[rep.slotIn]
			rep.dmut.RUnlock()
			_, skipSlot = rep.skipSlots[rep.slotIn]
		}
		// Found an empty slot
		delete(rep.requests, k)
		prop := &proposal{rep.slotIn, req}
		rep.pmut.Lock()
		rep.proposals[prop.hash()] = prop
		rep.pmut.Unlock()
		// Forward proposal to leader thread
		rep.leader.proposeInChan <- *prop
		rep.slotIn++
	}
	rep.rmut.Unlock()
}

// Perform method in Fig 1 of PMMC
func (rep *ReplicaAgent) perform(req *request) {
	rep.debugPrintf("performing %v\n", *req)
	rep.dmut.RLock()
	for s, oldDec := range rep.decisions {
		// If req has been previously committed, ignore it
		if rep.slotOut > s && oldDec.eq(req) {
			rep.slotOut++
			rep.dmut.RUnlock()
			return
		}
	}
	rep.dmut.RUnlock()
	// Else execute the request and perform output commit to client
	// "committed <clientID> <reqNum>."
	rep.chatLog = append(rep.chatLog, fmt.Sprintf("%d: %s", req.clientID, req.payload))
	response := fmt.Sprintf("committed %d %d", req.clientID, req.reqNum)
	rep.send(req.clientID, response)
	rep.slotOut++
	rep.debugPrintf("Commited {%d, %d, %s}\n", req.clientID, req.reqNum, req.payload)
}

// Handles a decision message "decision <slot> <clientID> <reqNum> <m>"
func (rep *ReplicaAgent) handleDecision(d string) {
	// Store decision in rep.decisions
	dSlice := strings.SplitN(d, " ", 5)
	slot, _ := strconv.ParseUint(dSlice[1], 10, 64)
	id, _ := strconv.ParseUint(dSlice[2], 10, 64)
	cid := c.ProcessID(id) // client who issued the request
	reqNum, _ := strconv.ParseUint(dSlice[3], 10, 64)
	m := dSlice[4]
	rep.debugPrintf("Received decision for %d : (%d, %d)\n", slot, cid, reqNum)
	newDec := &request{cid, reqNum, m}

	// ignore if decision already received
	rep.dmut.RLock()
	if _, ok := rep.decisions[slot]; ok {
		rep.dmut.RUnlock()
		return
	}
	rep.dmut.RUnlock()

	rep.dmut.Lock()
	rep.decisions[slot] = newDec
	rep.dmut.Unlock()

	// Execute all decisions that can be committed
	rep.dmut.RLock()
	decToExec, ok := rep.decisions[rep.slotOut]
	rep.dmut.RUnlock()
	for ok {
		rep.debugPrintf("HEYHEY for %d : (%d, %d)\n", slot, cid, reqNum)
		// If slot of request I am about to excute is used in proposals, then
		// 1. remove it from proposals, and
		// 2. if the req removed is not the one I am about to execute, put it back
		//    into rep.requests
		rep.pmut.Lock()
		for k, prop := range rep.proposals {
			if prop.slot == rep.slotOut {
				// If slotOut used for a command in rep.proposals
				delete(rep.proposals, k)
				if !prop.req.eq(decToExec) {
					// If req removed from rep.proposals is not decToExec
					rep.rmut.Lock()
					rep.requests[prop.hash()] = prop.req
					rep.rmut.Unlock()
				}
				break // No need to keep searching
			}
		}
		rep.pmut.Unlock()
		rep.perform(decToExec)
		rep.dmut.RLock()
		decToExec, ok = rep.decisions[rep.slotOut]
		rep.dmut.RUnlock()
	}
	if _, ok := rep.failureDetector.leaders[rep.myID]; ok {
		// propose() iff I am leader
		rep.propose()
	}
}
