package server

// This file contains the definition and main logic of an Ovid server.
// In particular, this is the central location where actions are
// triggered by incoming messages.

import (
	"fmt"
	"os"
	"runtime/debug"
	"sort"
	"strings"
	"time"

	c "github.com/TonyZhangND/GoOvid/commons"
	a "github.com/TonyZhangND/GoOvid/server/agents"
)

var (
	masterIP   string
	masterPort c.PortNum
	gridConfig map[c.ProcessID]*a.AgentInfo
	myBoxID    c.BoxID
	myAgents   map[c.ProcessID]*a.Agent
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
	destBox := gridConfig[phyDest].Box
	s := fmt.Sprintf("%d %d %d %s", senderID, phyDest, destPort, msg)
	linkMgr.send(destBox, s)
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
		os.Exit(0)
	default:
		debugPrintf("Invalid command %v from master\n", command)
	}
}

// Handles messages from a server
func handleServerMsg(data string) {
	msgLog.appendMsg(data)
}

// Helper: generates a slice containing all boxes in this configuration
func getAllBoxes() []c.BoxID {
	if gridConfig == nil {
		c.FatalOvidErrorf("grid cofiguration not initialized\n")
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
			var ag a.Agent
			switch agentInfo.Type {
			case a.Chat:
				ag = &a.ChatAgent{}
			default:
				c.FatalOvidErrorf("Invalid agent type for agent %v:%v\n", k, *agentInfo)
			}
			myAg[k] = &ag
		}
	}

	// Initialize and run each agent on this box
	for agentID, agent := range myAg {
		// Create custom send() func using closure
		sendMsg := func(vDest c.ProcessID, msg string) {
			phyDest := gridConfig[agentID].Routes[vDest].DestID
			destPort := gridConfig[agentID].Routes[vDest].DestPort
			send(agentID, phyDest, destPort, msg)
		}

		// Create custom error() func using closure
		fatalAgentErrorf := func(s string, a ...interface{}) {
			errMsg := fmt.Sprintf(s, a...)
			fmt.Printf("Error : Ovid : %s", errMsg)
			debug.PrintStack()
			(*agent).Halt()
		}
		// Initialize the agent
		(*agent).Init(gridConfig[agentID].RawAttrs, sendMsg, fatalAgentErrorf)
	}
	return myAg
}

// InitAndRunServer is the main method of a server
func InitAndRunServer(boxID c.BoxID, config map[c.ProcessID]*a.AgentInfo, mstrPort c.PortNum) {
	// Check for illegal values
	if mstrPort < 1024 {
		fmt.Printf("Port number %d is a well-known port and cannot be used "+
			"as masterPort\n", masterPort)
		os.Exit(1)
	}
	if mstrPort < 10000 {
		fmt.Printf("Port number %d is reserved for inter-server use\n",
			masterPort)
		os.Exit(1)
	}

	// Populate the global variables and start the linkManager
	gridConfig = config
	myBoxID = boxID
	masterIP = "127.0.0.1"
	masterPort = mstrPort
	shouldRun = true
	serverInChan := make(chan string) // used to receive inter-server messages
	masterInChan := make(chan string) // used to receive messages from the master
	linkMgr = newLinkManager(getAllBoxes(), serverInChan, masterInChan)
	msgLog = newMessageLog()
	debugPrintf("Launching server...\n")
	linkMgr.run()
	time.Sleep(1 * time.Second)
	debugPrintf(serverInfo())

	// Initialize and run my agents
	myAgents = initAgents()
	for _, agent := range myAgents {
		(*agent).Run()
	}

	// main loop
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
	for shouldRun {
		handleServerMsg(<-serverInChan)
	}
	debugPrintf("Terminating\n")
}
