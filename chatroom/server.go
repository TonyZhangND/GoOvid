package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	myPhysID   processID
	gridSize   uint16
	masterIP   string
	masterPort uint16
	gridIP     string
	shouldRun  bool
	masterConn net.Conn
	// a set of all known servers and their perceived status
	connStatus connTracker
	messageLog []string
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
	connStatus = newConnTracker(knownProcesses)
	messageLog = make([]string, 0, 100)
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
	aliveSet := connStatus.getAlive()
	rep := make([]string, 0)
	for _, pid := range aliveSet { // find the nodes that are up
		rep = append(rep, strconv.Itoa(int(pid)))
	}
	// compose and send response to master
	reply := "alive " + strings.Join(rep, ",")
	sendToMaster(reply)
}

// Responds to "get" command from the master
func doGet() {
	response := "messages " + strings.Join(messageLog, ",")
	sendToMaster(response)
}

// Responds to "broadcast" command from the master
func doBroadcast() {
	// TODO
}

// Dials for new connections to all pid <= my pid
func dialForConnections() {
	for shouldRun {
		down := connStatus.getDead()
		for _, pid := range down {
			if pid <= myPhysID && !connStatus.isUp(pid) {
				dialingAddr := fmt.Sprintf("%s:%d", gridIP, basePort+pid)
				c, err := net.DialTimeout("tcp", dialingAddr,
					20*time.Millisecond)
				if err == nil {
					connStatus.markAsUp(pid, c)
					go handleConnection(c)
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
		go handleConnection(c)
	}
}

// TODO: Main thread for each server connection
func handleConnection(conn net.Conn) {
	// TODO: Spawn a failure detector
	defer conn.Close()
	fmt.Printf("Serving %s\n", conn.RemoteAddr().String())
	for shouldRun {
		data, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Println(err)
		}
		dataSlice := strings.SplitN(strings.TrimSpace(data), " ", 2)
		header := dataSlice[0]
		payload := dataSlice[1]
		fmt.Printf("Type of message: %v\n", header)
		switch header {
		case "ping":
			doRcvPing(payload, conn)
		case "msg":
			doRcvMsg(payload)
		default:
			fmt.Printf("Error, Invalid msg %v from master\n", header)
		}
		fmt.Println("Done responding to master")
		time.Sleep(2 * time.Second)
	}
}

func doRcvPing(s string, conn net.Conn) {
	sender, err := strconv.Atoi(s)
	if err != nil {
		fmt.Printf("Error, Invalid ping received by %v\n", myPhysID)
		os.Exit(1)
	}
	// check that sender is a valid process
	if !connStatus.isKnown(processID(sender)) {
		fmt.Printf("Error, Invalid ping from  %v received by %v\n",
			myPhysID, sender)
		os.Exit(1)
	}
	connStatus.markAsUp(processID(sender), conn)
}

func doRcvMsg(s string) {
	// sSlice := strings.SplitN(strings.TrimSpace(s), " ", 2)
	// sender := sSlice[0]
	// msg := sSlice[1]
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
	if masterPort > 2999 {
		fmt.Printf("Port number %d is reserved for inter-server use\n", masterPort)
		os.Exit(1)
	}

	// initialize server
	fmt.Println("Launching server...")
	initServer(processID(pid), uint16(gridSize), uint16(masterPort))
	fmt.Println(serverInfo())

	// listen for master on the master address
	masterAddr := fmt.Sprintf("%s:%d", masterIP, masterPort)
	fmt.Println("Listening for master connecting on " + masterAddr)
	mstrListener, _ := net.Listen("tcp", masterAddr)
	mstrConn, _ := mstrListener.Accept()
	defer mstrConn.Close()
	masterConn = mstrConn
	fmt.Println("Accepted master connection")

	// initialize and maintain connections with peers
	fmt.Println("Listening for peer connections")
	go listenForConnections()
	fmt.Println("Dialing for peer connections")
	go dialForConnections()

	// main loop: process commands from master
	for shouldRun {
		data, err := bufio.NewReader(masterConn).ReadString('\n')
		if err != nil {
			fmt.Println(err)
		}
		command := strings.TrimSpace(data)
		fmt.Printf("Command from master: %v\n", command)
		switch command {
		case "get":
			doGet()
		case "alive":
			doAlive()
		case "broadcast":
			doBroadcast()
		default:
			fmt.Printf("Error, invalid command %v from master\n", command)
		}
		fmt.Println("Done responding to master")
	}
	fmt.Println("Terminating")
	os.Exit(0)
}
