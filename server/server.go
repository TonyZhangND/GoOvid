package server

// This file contains the definition and main logic of an Ovid server.
// In particular, this is the central location where actions are
// triggered by incoming messages.

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	c "github.com/TonyZhangND/GoOvid/commons"
)

var (
	myPhysID   c.ProcessID
	gridSize   uint16
	masterIP   string
	masterPort c.PortNum
	gridIP     string
	shouldRun  bool // loop condition for the server's routines
	linkMgr    *linkManager
	msgLog     *messageLog
)

// Populates the global variables and starts the linkManager
func initAndRunServer(pid c.ProcessID, gridSz uint16,
	mstrPort c.PortNum, serverInChan chan string, masterInChan chan string) {
	myPhysID = pid
	gridSize = gridSz
	masterIP = "127.0.0.1"
	masterPort = mstrPort
	gridIP = "127.0.0.1"
	shouldRun = true
	knownProcesses := make([]c.ProcessID, gridSz)
	for i := 0; i < int(gridSz); i++ {
		knownProcesses[i] = c.ProcessID(i)
	}
	linkMgr = newLinkManager(knownProcesses, serverInChan, masterInChan)
	msgLog = newMessageLog()
	linkMgr.run()
}

// Returns a string describing this server
func serverInfo() string {
	return fmt.Sprintf("* GoOvid server *\n"+
		"physID: %d\n"+
		"gridSize: %d\n"+
		"masterPort: %d\n",
		myPhysID, gridSize, masterPort)
}

// Responds to an "alive" command from the master
func doAlive() {
	aliveSet := linkMgr.getAllUp()
	sort.Slice(aliveSet,
		func(i, j int) bool { return aliveSet[i] < aliveSet[j] })
	rep := make([]string, len(aliveSet))
	for i, pid := range aliveSet { // find the nodes that are up
		rep[i] = strconv.Itoa(int(pid))
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
		msg := fmt.Sprintf("Invalid command %v from master", command)
		debugPrintln(msg)
	}
}

// Handles messages from a server
func handleServerMsg(data string) {
	msgLog.appendMsg(data)
}

func main() {
	// process command line arguments
	pid, err1 := strconv.ParseUint(os.Args[1], 10, 16)
	gridSize, err2 := strconv.ParseUint(os.Args[2], 10, 16)
	masterPort, err3 := strconv.ParseUint(os.Args[3], 10, 16)
	errMsg := fmt.Sprintf("Errors occured while processing arguments.\n"+
		"PhysID: %v\n"+
		"gridSize: %v\n"+
		"masterPort: %v\n"+
		"Program exiting...\n",
		err1, err2, err3)
	checkFatalServerError(err1, errMsg)
	checkFatalServerError(err2, errMsg)
	checkFatalServerError(err3, errMsg)
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

	// initialize server
	debugPrintln("Launching server...")
	serverInChan := make(chan string) // used to receive inter-server messages
	masterInChan := make(chan string) // used to receive messages from the master
	initAndRunServer(c.ProcessID(pid), uint16(gridSize), c.PortNum(masterPort),
		serverInChan, masterInChan)
	debugPrintln(serverInfo())

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
	debugPrintln("Terminating")
}
