package server

// This file contains the definition and main logic of an Ovid server.
// In particular, this is the central location where actions are
// triggered by incoming messages.

import (
	"fmt"
	"math/rand"
	"os"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"time"

	a "github.com/TonyZhangND/GoOvid/agents"
	c "github.com/TonyZhangND/GoOvid/commons"
)

var (
	masterIP   string
	masterPort c.PortNum
	gridConfig map[c.ProcessID]*a.AgentInfo
	myBoxID    c.BoxID
	myAgents   map[c.ProcessID]*a.Agent
	lossRate   float64
	shouldRun  bool // loop condition for the server's routines
	linkMgr    *linkManager
	msgLog     *messageLog
)

// Returns a string describing this server
func serverInfo() string {
	return fmt.Sprintf("* GoOvid server *\n"+
		"myBoxID: %s\n"+
		"Boxes in grid: %v\n"+
		"masterPort: %d\n",
		myBoxID, linkMgr.getAllKnown(), masterPort)
}

// Sends a message to phyDest
func send(senderID, phyDest c.ProcessID, destPort c.PortNum, msg string) {
	// Check destination is valid
	destAgent, ok := gridConfig[phyDest]
	if !ok {
		fatalServerErrorf("Destination agent %v does not exist\n", phyDest)
	}

	// Drop message according to loss rate
	rand.Seed(time.Now().UnixNano())
	if rand.Float64() < lossRate {
		return
	}

	// Send the message
	destBox := destAgent.Box
	if destBox == myBoxID {
		// if sending to agent on this box
		(*myAgents[phyDest]).Deliver(msg, destPort)
	} else {
		// else sending to agent on some other box
		s := fmt.Sprintf("%d %d %d %s", senderID, phyDest, destPort, msg)
		linkMgr.send(destBox, s)
	}
}

// Responds to an "alive" command from the master
func doAlive() {
	aliveSet := linkMgr.getAllUp()
	sort.Slice(aliveSet,
		func(i, j int) bool { return aliveSet[i] < aliveSet[j] })
	rep := make([]string, len(aliveSet))
	for i, bid := range aliveSet { // find the nodes that are up
		rep[i] = string(bid)
	}
	// compose and send response to master
	reply := "alive " + strings.Join(rep, ",")
	linkMgr.sendToMaster(reply)
}

// Responds to "get" command from the master
func doGet() {
	response := "messages " + strings.Join(msgLog.getMessages(), ",")
	linkMgr.sendToMaster(response)
}

// Responds to "broadcast" command from the master
func doBroadcast(msg string) {
	linkMgr.broadcast(msg)
}

// Handles messages from the master
func handleMasterMsg(data string) {
	dataSlice := strings.SplitN(strings.TrimSpace(data), " ", 2)
	command := dataSlice[0]
	switch command {
	case "get":
		doGet()
	case "alive":
		doAlive()
	case "broadcast":
		payload := dataSlice[1]
		doBroadcast(payload)
	case "crash":
		// self-destruct
		for _, agent := range myAgents {
			(*agent).Halt()
		}
		shouldRun = false
		os.Exit(0)
	default:
		debugPrintf("Invalid command %v from master\n", command)
	}
}

// Handles messages from a server
func handleServerMsg(data string) {
	if strings.SplitN(data, " ", 2)[0] == "chatroom" {
		// if for chatroom project
		dataSlice := strings.SplitN(strings.TrimSpace(data), " ", 3)
		// senderBox := dataSlice[1]
		msgLog.appendMsg(dataSlice[2])
	} else {
		// else a GoOvid message to deliver to an agent
		dataSlice := strings.SplitN(data, " ", 4)
		_, err := strconv.ParseInt(dataSlice[0], 10, 16) //ignore sender for now
		checkFatalServerErrorf(err, "Cannot parse sender of incoming message '%s'\n", data)
		destID, err := strconv.ParseInt(dataSlice[1], 10, 16)
		checkFatalServerErrorf(err, "Cannot parse destID of incoming message '%s'\n", data)
		destPort, err := strconv.ParseInt(dataSlice[2], 10, 16)
		checkFatalServerErrorf(err, "Cannot parse destPort of incoming message '%s'\n", data)
		(*myAgents[c.ProcessID(destID)]).Deliver(dataSlice[3], c.PortNum(destPort))
	}
}

// Helper: generates a slice containing all boxes in this configuration
func getAllBoxes() []c.BoxID {
	if gridConfig == nil {
		c.FatalOvidErrorf("Grid cofiguration not initialized\n")
	}
	boxSet := make(map[c.BoxID]int)
	for _, agentInfo := range gridConfig {
		boxSet[agentInfo.Box] = 1
	}
	boxes := make([]c.BoxID, len(boxSet))
	i := 0
	for bid := range boxSet {
		boxes[i] = bid
		i++
	}
	return boxes
}

// Helper: initializes all agents on this box
func initAgents() map[c.ProcessID]*a.Agent {
	if gridConfig == nil {
		c.FatalOvidErrorf("grid cofiguration not initialized\n")
	}
	// Make map containing all agent structs on this box
	myAg := make(map[c.ProcessID]*a.Agent)
	for k, agentInfo := range gridConfig {
		if agentInfo.Box == myBoxID {
			// allocate the struct
			ag := a.NewAgent(agentInfo.Type)
			myAg[k] = &ag
		}
	}
	// Initialize and run each agent on this box
	for agentID, agent := range myAg {
		// Notice that for each of the following closures, we are implementing a closure
		// generator rather than a closure itself. This is because in Go, variables
		// declared in for loops are passed by reference. In other words, bad
		// things happen when you use a loop variable in the closure, because those values
		// can change from underneath you, and the closure is then messed up.
		// However, fuction params are passed by value. Thus, we use this generator
		// technique to "freeze" the agentID variable for each closure, for each agent.

		// Create custom send func using closure
		sendFuncGen := func(id c.ProcessID) func(vDest c.ProcessID, msg string) {
			return func(vDest c.ProcessID, msg string) {
				phyDest := gridConfig[id].Routes[vDest].DestID
				destPort := gridConfig[id].Routes[vDest].DestPort
				send(id, phyDest, destPort, msg)
			}
		}
		// Create custom error func using closure
		fatalAgentErrorfGen := func(id c.ProcessID) func(s string, a ...interface{}) {
			return func(s string, a ...interface{}) {
				errMsg := fmt.Sprintf(s, a...)
				fmt.Printf("Error : Agent %v : %s", id, errMsg)
				debug.PrintStack()
				(*agent).Halt()
			}
		}
		// Create custom debugPrintf func using closure
		agentDebugPrintfGen := func(id c.ProcessID) func(s string, a ...interface{}) {
			return func(s string, a ...interface{}) {
				msg := fmt.Sprintf("Agent %v : %s", id, s)
				debugPrintf(msg, a...)
			}
		}
		// Initialize the agent
		(*agent).Init(gridConfig[agentID].RawAttrs,
			sendFuncGen(agentID),
			fatalAgentErrorfGen(agentID),
			agentDebugPrintfGen(agentID))
	}
	return myAg
}

// InitAndRunServer is the main method of a server
func InitAndRunServer(
	boxID c.BoxID,
	config map[c.ProcessID]*a.AgentInfo,
	mstrPort c.PortNum, // 0 if master conn not specified
	loss float64) {

	// Check for illegal values
	if mstrPort != 0 {
		if mstrPort < 1024 {
			fmt.Printf("Port number %d is a well-known port and cannot be used "+
				"for master connection\n", mstrPort)
			os.Exit(1)
		}
		if mstrPort < 10000 {
			fmt.Printf("Port number %d is reserved for inter-server use\n",
				masterPort)
			os.Exit(1)
		}
	}

	// Populate the global variables and start the linkManager
	gridConfig = config
	myBoxID = boxID
	masterIP = "127.0.0.1"
	masterPort = mstrPort
	lossRate = loss
	shouldRun = true
	serverInChan := make(chan string) // used to receive inter-server messages
	masterInChan := make(chan string) // used to receive messages from the master
	linkMgr = newLinkManager(
		getAllBoxes(),
		serverInChan,
		masterInChan)
	msgLog = newMessageLog()
	debugPrintf("Launching server...\n")
	linkMgr.run()
	time.Sleep(1 * time.Second)
	debugPrintf(serverInfo())

	// Initialize my agents
	myAgents = initAgents()

	// main loop
	if masterPort > 0 {
		// only listen to master if master port specified
		go func() {
			// There is an important reason why this is a separate goroutine,
			// rather than within a select block together with serverInChan.
			// Because a broadcast includes pushing into serverInChan,
			// handleMasterMessage may block, resulting in a deadlock. In fact,
			// while a buffered channel can defer such a deadlock, the deadlock
			// will inevitably remain a reachable execution. The solution is
			// what I have here -- decouple the synchrony between the two channels.
			// Naively, one could do `go handleMasterMsg(<-masterInChan)`,
			// but that breaks FIFO ordering
			for shouldRun {
				handleMasterMsg(<-masterInChan)
			}
		}()
	}
	// run my agents
	for _, agent := range myAgents {
		go (*agent).Run()
	}
	for shouldRun {
		handleServerMsg(<-serverInChan)
	}
	debugPrintf("Terminating\n")
}
