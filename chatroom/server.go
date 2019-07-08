package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

var (
	physID     processID
	gridSize   uint16
	masterIP   string
	masterPort uint16
	gridIP     string
	shouldRun  bool
	masterConn net.Conn
	// a set of all known servers and their perceived status
	failureDetector stateTracker
	messageLog      []string
)

// newServer is the constructor for server.
// It returns a server struct with default values for some fields.
func initServer(pid processID, gridSz uint16, mstrPort uint16) {
	physID = pid
	gridSize = gridSz
	masterIP = "127.0.0.1"
	masterPort = mstrPort
	gridIP = "127.0.0.1"
	shouldRun = true
	masterConn = nil
	failureDetector = newStateTracker()
	messageLog = make([]string, 0, 100)
	// populate failure detector with all nodes
	for pid := uint16(0); pid < gridSz; pid++ {
		failureDetector.addProcess(processID(pid))
	}
}

// String is the "toString" method for this server
// It returns a string describing this server
func serverInfo() string {
	return fmt.Sprintf("* GoOvid server *\n"+
		"physID: %d\n"+
		"gridSize: %d\n"+
		"masterPort: %d\n",
		physID, gridSize, masterPort)
}

// sendToMaster sends msg string to the master
func sendToMaster(msg string) {
	_, err := masterConn.Write([]byte(msg + "\n"))
	if err != nil {
		fmt.Printf("Error occured while sending msg '%v' to master: %v",
			msg, err)
	}
}

// Respond to an "alive" command from the master
func doAlive() {
	aliveSet := failureDetector.getAlive()
	response := make([]string, 0)
	for _, pid := range aliveSet { // find the nodes that are up
		response = append(response, string(pid))
	}
	// compose and send response to master
	reply := "alive " + strings.Join(response, ",")
	sendToMaster(reply)
}

// Respond to "get" command from the master
func doGet() {
	response := "messages " + strings.Join(messageLog, ",")
	sendToMaster(response)
}

// Respond to "broadcast" command from the master
func doBroadcast() {
	// TODO
}

// Listen and establish new connections
func listenForConnections(l net.Listener) {
	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		go handleConnection(c)
	}
}

// Main thread for each server connection
func handleConnection(c net.Conn) {
	// TODO: Spawn a failure detector
	fmt.Printf("Serving %s\n", c.RemoteAddr().String())
	// for {
	// 	data, err := bufio.NewReader(c).ReadString('\n')
	// 	if err != nil {
	// 		fmt.Println(err)
	// 		return
	// 	}
	// 	msg := strings.TrimSpace(string(data))
	// 	if msg == "gekkuP" {
	// 		break
	// 	}
	// 	c.Write([]byte("hello"))
	// }
	c.Close()
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
	fmt.Println("Accepted master connection. Listening for input...")

	// initialize and maintain connections with peers
	l, err := net.Listen("tcp4", string(basePort+physID))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer l.Close()
	go listenForConnections(l)

	// process inputs from master
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
			fmt.Printf("Invalid command %v from master\n", command)
		}
		fmt.Println("Done responding to master")
	}
	fmt.Println("Terminating")
	os.Exit(0)
}
