package server

// This file contains the definition and main logic of an Ovid server.
// In particular, this is the central location where actions are
// triggered by incoming messages.

import (
	"fmt"
	"os"
	"sort"
	"strings"

	c "github.com/TonyZhangND/GoOvid/commons"
)

var (
	myBoxID    c.BoxID
	masterIP   string
	masterPort c.PortNum
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

// InitAndRunServer is the main method of a server
func InitAndRunServer(boxID c.BoxID, knownProcesses []c.BoxID, mstrPort c.PortNum) {

	if masterPort < 1024 {
		fmt.Printf("Port number %d is a well-known port and cannot be used "+
			"as masterPort\n", masterPort)
		os.Exit(1)
	}
	if masterPort < 10000 {
		fmt.Printf("Port number %d is reserved for inter-server use\n",
			masterPort)
		os.Exit(1)
	}

	// Populate the global variables and starts the linkManager
	myBoxID = boxID
	masterIP = "127.0.0.1"
	masterPort = mstrPort
	shouldRun = true

	serverInChan := make(chan string) // used to receive inter-server messages
	masterInChan := make(chan string) // used to receive messages from the master
	linkMgr = newLinkManager(knownProcesses, serverInChan, masterInChan)
	msgLog = newMessageLog()
	debugPrintf("Launching server...\n")
	linkMgr.run()
	debugPrintf(serverInfo())

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
