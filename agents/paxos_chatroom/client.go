package paxos

// This file contains the definition and logic of a client using the paxos service.
// The ClientAgent type must implement the Agent interface.

import (
	"fmt"
	"strings"
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
	log      string // TODO: This doesn't do anything now

	// Client state
	nextReqNum uint64
	reqQueue   []*req
}

// req struct represents a client request
type req struct {
	reqNum uint64
	m      string
	ticker *time.Ticker // used to mark intervals after which request should be re-issued
	done   chan bool    // used to stop ticker when request is committed
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
	clt.log = attrs["log"].(string)

	// Initialize client statee
	clt.nextReqNum = 0
	clt.reqQueue = make([]*req, 0)
}

// Halt stops the execution of the agent.
func (clt *ClientAgent) Halt() {
	clt.isActive = false
}

// Deliver a message
func (clt *ClientAgent) Deliver(request string, port c.PortNum) {
	switch port {
	case 9:
		// Receive msg "issue <m>"
		msgSlice := strings.SplitN(request, " ", 2)
		if msgSlice[0] != "issue" {
			clt.fatalAgentErrorf(
				"Received unexpected command '%s' in unexpected port %v",
				request, port)
		}
		// Append request to reqQueue
		r := &req{
			reqNum: clt.nextReqNum,
			m:      msgSlice[1],
			done:   make(chan bool)}
		clt.reqQueue = append(clt.reqQueue, r)
	default:
		clt.fatalAgentErrorf("Received '%s' in unexpected port %v", request, port)
	}
}

// Run begins the execution of the paxos agent.
func (clt *ClientAgent) Run() {
	clt.isActive = true
	go clt.mainThread()
	wg.Wait()
}

// Main execution thread of client agent
func (clt *ClientAgent) mainThread() {
	wg.Add(1)
	defer wg.Done()
	for clt.isActive {
		if len(clt.reqQueue) == 0 {
			time.Sleep(sleepDuration)
		} else { // Process first message in queue
			r := clt.reqQueue[0] // Outstanding request

			// Broadcast request
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
				case <-r.ticker.C:
					// Timer expired, resend request
					clt.debugPrintf("Timer expired\n")
					for rep := range clt.replicas {
						clt.send(rep, fmt.Sprintf("%d %d %s", clt.myID, r.reqNum, r.m))
					}
				}
			}
			clt.reqQueue = clt.reqQueue[1:]
		}
	}
}
