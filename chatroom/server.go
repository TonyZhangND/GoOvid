package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	debug      = false
	myPhysID   processID
	gridSize   uint16
	masterIP   string
	masterPort uint16
	gridIP     string
	shouldRun  bool
	masterConn net.Conn
	// a set of all known servers and their perceived status
	linkMgr *linkManager
	msgLog  *messageLog
)

// newServer is the constructor for server.
// It returns a server struct with default values for some fields.
func initServer(pid processID, gridSz uint16, mstrPort uint16) {
	myPhysID = pid
	gridSize = gridSz
	masterIP = "127.0.0.1"
	masterPort = mstrPort
	gridIP = "127.0.0.1"
	shouldRun = true
	masterConn = nil
	knownProcesses := make([]processID, gridSz)
	for i := 0; i < int(gridSz); i++ {
		knownProcesses[i] = processID(i)
	}
	linkMgr = newLinkManager(knownProcesses)
	msgLog = newMessageLog()
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

// sendToMaster sends msg string to the master
func sendToMaster(msg string) {
	_, err := masterConn.Write([]byte(msg + "\n"))
	if err != nil {
		fmt.Printf("Error occured while sending msg '%v' to master: %v",
			msg, err)
	}
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
	sendToMaster(reply)
}

// Responds to "get" command from the master
func doGet() {
	response := "messages " + strings.Join(msgLog.getMessages(), ",")
	sendToMaster(response)
}

// Responds to "broadcast" command from the master
func doBroadcast(msg string) {
	linkMgr.broadcast(msg)
}

// Dials for new connections to all pid <= my pid
func dialForConnections() {
	for shouldRun {
		down := linkMgr.getDead()
		for _, pid := range down {
			if pid <= myPhysID && !linkMgr.isUp(pid) {
				dialingAddr := fmt.Sprintf("%s:%d", gridIP, basePort+pid)
				c, err := net.DialTimeout("tcp", dialingAddr,
					20*time.Millisecond)
				if err == nil {
					l := newLinkKnownOther(c, pid)
					go l.handleConnection()
				}
			}
		}
		time.Sleep(1 * time.Second)
	}
}

// Listens and establishes new connections
func listenForConnections() {
	listenerAddr := fmt.Sprintf("%s:%d", gridIP, basePort+myPhysID)
	l, err := net.Listen("tcp", listenerAddr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer l.Close()
	for shouldRun {
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		l := newLink(c)
		go l.handleConnection()
	}
}

func main() {
	// process command line arguments
	pid, err1 := strconv.ParseUint(os.Args[1], 10, 16)
	gridSize, err2 := strconv.ParseUint(os.Args[2], 10, 16)
	masterPort, err3 := strconv.ParseUint(os.Args[3], 10, 16)
	if err1 != nil || err2 != nil || err3 != nil {
		fmt.Printf("Errors occured while processing arguments.\n"+
			"PhysID: %v\n"+
			"gridSize: %v\n"+
			"masterPort: %v\n"+
			"Program exiting...\n",
			err1, err2, err3)
		os.Exit(1)
	}
	if masterPort < 1024 {
		fmt.Printf("Port number %d is a well-known port and cannot be used "+
			"as masterPort\n", masterPort)
		os.Exit(1)
	}
	if masterPort < 10000 {
		fmt.Printf("Port number %d is reserved for inter-server use\n", masterPort)
		os.Exit(1)
	}

	// initialize server
	debugPrintln("Launching server...")
	initServer(processID(pid), uint16(gridSize), uint16(masterPort))
	debugPrintln(serverInfo())

	// initialize and maintain connections with peers
	debugPrintln("Listening for peer connections")
	go listenForConnections()
	debugPrintln("Dialing for peer connections")
	go dialForConnections()

	// listen for master on the master address
	masterAddr := fmt.Sprintf("%s:%d", masterIP, masterPort)
	debugPrintln("Listening for master connecting on " + masterAddr)
	mstrListener, _ := net.Listen("tcp", masterAddr)
	mstrConn, _ := mstrListener.Accept()
	defer mstrConn.Close()
	masterConn = mstrConn
	debugPrintln("Accepted master connection")

	// main loop: process commands from master
	for shouldRun {
		data, err := bufio.NewReader(masterConn).ReadString('\n')
		if err != nil {
			fmt.Println(err)
		}
		if data == "" {
			// the connection is dead. Kill this server
			shouldRun = false
			break
		}
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
			masterConn.Close()
			os.Exit(0)
		default:
			fmt.Printf("Error, invalid command %v from master\n", command)
		}
	}
	debugPrintln("Terminating")
}
