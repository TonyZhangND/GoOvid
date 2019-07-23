package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

var (
	debugMode  = true
	myPhysID   processID
	gridSize   uint16
	masterIP   string
	masterPort uint16
	gridIP     string
	shouldRun  bool
	// a set of all known servers and their perceived status
	linkMgr *linkManager
	msgLog  *messageLog
)

// newServer is the constructor for server.
// It returns a server struct with default values for some fields.
func initAndRunServer(pid processID, gridSz uint16,
	mstrPort uint16, serverInChan chan string, masterInChan chan string) {
	myPhysID = pid
	gridSize = gridSz
	masterIP = "127.0.0.1"
	masterPort = mstrPort
	gridIP = "127.0.0.1"
	shouldRun = true
	knownProcesses := make([]processID, gridSz)
	for i := 0; i < int(gridSz); i++ {
		knownProcesses[i] = processID(i)
	}
	linkMgr = newLinkManager(knownProcesses, serverInChan, masterInChan)
	msgLog = newMessageLog()
	linkMgr.run()
}

// String is the "toString" method for this server
// It returns a string describing this server
func serverInfo() string {
	return fmt.Sprintf("* GoOvid server *\n"+
		"physID: %d\n"+
		"gridSize: %d\n"+
		"masterPort: %d\n",
		myPhysID, gridSize, masterPort)
}

// Responds to an "alive" command from the master
func doAlive() {
	aliveSet := linkMgr.getAlive()
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

func handleServerMsg(data string) {
	msgLog.appendMsg(data)
}

func main() {
	// process command line arguments
	pid, err1 := strconv.ParseUint(os.Args[1], 10, 16)
	gridSize, err2 := strconv.ParseUint(os.Args[2], 10, 16)
	masterPort, err3 := strconv.ParseUint(os.Args[3], 10, 16)
	if err1 != nil || err2 != nil || err3 != nil {
		errMsg := fmt.Sprintf("Errors occured while processing arguments.\n"+
			"PhysID: %v\n"+
			"gridSize: %v\n"+
			"masterPort: %v\n"+
			"Program exiting...\n",
			err1, err2, err3)
		fatalError(errMsg)
	}
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
	initAndRunServer(processID(pid), uint16(gridSize), uint16(masterPort),
		serverInChan, masterInChan)
	debugPrintln(serverInfo())

	for shouldRun {
		select {
		case masterData := <-masterInChan:
			handleMasterMsg(masterData)
		case serverData := <-serverInChan:
			handleServerMsg(serverData)
		}
	}
	debugPrintln("Terminating")
}
