package paxos

// This file contains the definition and logic of a client using the paxos service.
// The ClientAgent type must implement the Agent interface.

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	c "github.com/TonyZhangND/GoOvid/commons"
)

// ClientAgent struct contains the information inherent to a paxos client
type ClientAgent struct {
	// Default GoOvid states
	send             func(vDest c.ProcessID, msg string)
	fatalAgentErrorf func(errMsg string, a ...interface{})
	debugPrintf      func(s string, a ...interface{})
	isActive         bool

	// Client attributes
	myID     c.ProcessID
	replicas map[c.ProcessID]int
	mode     string // script or manual modes

	// Client state
	nextReqNum uint64
	reqQueue   []*req
	qmut       *sync.RWMutex
	nmut       *sync.RWMutex
}

// req struct represents a client request
type req struct {
	reqNum          uint64
	m               string
	ticker          *time.Ticker // used to mark intervals after which req should be re-issued
	timeoutMultiple int
	done            chan bool // used to inform mainThread that req has been committed
}

// Init fills the empty client struct with this agent's fields and attributes.
func (clt *ClientAgent) Init(attrs map[string]interface{},
	send func(vDest c.ProcessID, msg string),
	fatalAgentErrorf func(errMsg string, a ...interface{}),
	debugPrintf func(s string, a ...interface{})) {
	clt.send = send
	clt.fatalAgentErrorf = fatalAgentErrorf
	clt.debugPrintf = debugPrintf
	clt.isActive = false

	// Initialize client attributes
	clt.myID = c.ProcessID(attrs["myid"].(float64))
	clt.replicas = make(map[c.ProcessID]int)
	for _, x := range attrs["replicas"].([]interface{}) {
		id := c.ProcessID(x.(float64))
		clt.replicas[id] = 0
	}
	clt.mode = attrs["mode"].(string)
	if clt.mode != "script" && clt.mode != "manual" {
		clt.fatalAgentErrorf("Invalid mode '%s'\n", clt.mode)
	}

	// Initialize client state
	clt.nextReqNum = 0
	clt.reqQueue = make([]*req, 0)
	clt.qmut = new(sync.RWMutex)
	clt.nmut = new(sync.RWMutex)
}

// Halt stops the execution of the agent.
func (clt *ClientAgent) Halt() {
	clt.isActive = false
}

// Deliver a message
func (clt *ClientAgent) Deliver(request string, port c.PortNum) {
	switch port {
	case 1: // incoming msg from replica
		msgSlice := strings.SplitN(request, " ", 3)
		if msgSlice[0] != "committed" {
			clt.fatalAgentErrorf(
				"Received unexpected command '%s' in port %v\n",
				request, port)
		}
		// Receive msg "committed <clientID> <reqNum>"
		id, _ := strconv.ParseUint(msgSlice[1], 10, 64)
		n, _ := strconv.ParseUint(msgSlice[2], 10, 64)
		if c.ProcessID(id) != clt.myID {
			clt.fatalAgentErrorf(
				"Received unexpected commit response '%s'\n", request)
		}

		clt.qmut.RLock()
		notEmpty := len(clt.reqQueue) > 0
		matchReq := false
		if notEmpty {
			matchReq = clt.reqQueue[0].reqNum == n
		}
		clt.qmut.RUnlock()
		if notEmpty && matchReq {
			// If this is a response to a currently outstanding request,
			// stop the ticker and declare the request as done
			clt.qmut.Lock()
			clt.reqQueue[0].ticker.Stop()
			clt.reqQueue[0].done <- true
			close(clt.reqQueue[0].done) // done with this request, close the channel
			clt.reqQueue = clt.reqQueue[1:]
			clt.qmut.Unlock()
		}

	case 9: // incoming msg from controller
		// Receive msg "issue <m>"
		msgSlice := strings.SplitN(request, " ", 2)
		if msgSlice[0] != "issue" {
			clt.fatalAgentErrorf(
				"Received unexpected command '%s' in unexpected port %v\n",
				request, port)
		}
		// Append request to reqQueue
		clt.nmut.RLock()
		r := &req{
			reqNum:          clt.nextReqNum,
			m:               msgSlice[1],
			done:            make(chan bool),
			timeoutMultiple: 1}
		clt.nmut.RUnlock()
		clt.qmut.Lock()
		clt.reqQueue = append(clt.reqQueue, r)
		clt.qmut.Unlock()
		clt.nmut.Lock()
		clt.nextReqNum++
		clt.nmut.Unlock()
	default:
		clt.fatalAgentErrorf("Received '%s' in unexpected port %v\n", request, port)
	}
}

// Run begins the execution of the paxos agent.
func (clt *ClientAgent) Run() {
	clt.isActive = true
	if clt.mode == "script" {
		go clt.runScriptMode()
	}
	clt.mainThread()
}

func (clt *ClientAgent) runScriptMode() {
	for clt.isActive {
		clt.nmut.RLock()
		r := &req{
			reqNum:          clt.nextReqNum,
			m:               fmt.Sprintf("(%d : %d)", clt.myID, clt.nextReqNum),
			done:            make(chan bool),
			timeoutMultiple: 1}
		clt.nmut.RUnlock()
		clt.qmut.Lock()
		l := len(clt.reqQueue)
		clt.qmut.Unlock()
		if l == 0 {
			clt.reqQueue = append(clt.reqQueue, r)
			continue
		}
		clt.nmut.Lock()
		clt.nextReqNum++
		clt.nmut.Unlock()
		time.Sleep(commandInterval)
	}
}

// Main execution thread of client agent
func (clt *ClientAgent) mainThread() {
	for clt.isActive {
		clt.qmut.RLock()
		isEmpty := len(clt.reqQueue) == 0
		clt.qmut.RUnlock()
		if isEmpty {
			// No pending requests. Take a break, have a KitKat
			time.Sleep(sleepDuration)
		} else {
			// Process first message in queue
			clt.qmut.RLock()
			r := clt.reqQueue[0] // Outstanding request
			clt.qmut.RUnlock()

			// Broadcast request "<clientID> <reqNum> <m>" to replicas
			clt.debugPrintf("ISSUE request %d : '%s'\n", r.reqNum, r.m)
			for rep := range clt.replicas {
				clt.send(rep, fmt.Sprintf("%d %d %s", clt.myID, r.reqNum, r.m))
			}

			// Wait for request to be committed
			r.ticker = time.NewTicker(timeoutDuration)
			committed := false
			for !committed {
				select {
				case <-r.done:
					// Request committed
					committed = true
					clt.debugPrintf("COMMIT request %d : '%s'\n",
						r.reqNum, r.m)
				case <-r.ticker.C:
					// Timer expired, resend request
					// increment timer duration
					r.timeoutMultiple++
					duration := timeoutDuration
					for i := 0; i < r.timeoutMultiple; i++ {
						duration = duration * 2
					}
					r.ticker = time.NewTicker(duration)
					clt.debugPrintf("Timeout for (%d, %d)\n", clt.myID, r.reqNum)
					for rep := range clt.replicas {
						clt.send(rep, fmt.Sprintf("%d %d %s", clt.myID, r.reqNum, r.m))
					}
				}
			}
		}
	}
}
