package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

type (
	status    int // the perceived operational status of a node
	processID uint16
)

const (
	alive = iota
	dead  = iota
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
	failureDetector map[processID]status
)

// // newServer is the constructor for server.
// // It returns a server struct with default values for some fields.
func initServer(pid processID, gridSz uint16, mstrPort uint16) {
	physID = pid
	gridSize = gridSz
	masterIP = "127.0.0.1"
	masterPort = mstrPort
	gridIP = "127.0.0.1"
	shouldRun = true
	masterConn = nil
	failureDetector = make(map[processID]status)
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

func doAlive() {
	// compute the set of live nodes
	aliveSet := make([]string, len(failureDetector))
	for physID, state := range failureDetector {
		if state == alive {
			aliveSet = append(aliveSet, string(physID))
		}
	}
	// compose and send response to master
	response := "alive " + strings.Join(aliveSet, ",")
	sendToMaster(response)
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

	// initialize server
	fmt.Println("Launching server...")
	initServer(
		processID(pid),
		uint16(gridSize),
		uint16(masterPort))
	fmt.Println(serverInfo())

	// listen for master on the master address
	masterAddr := fmt.Sprintf("%s:%d", masterIP, masterPort)
	fmt.Println("Listening for master connecting on " + masterAddr)
	mstrListener, _ := net.Listen("tcp", masterAddr)
	mstrConn, _ := mstrListener.Accept()
	defer mstrConn.Close()
	masterConn = mstrConn
	fmt.Println("Accepted master connection. Listening for input...")

	for shouldRun {
		// process inputs from master
		status, err := bufio.NewReader(masterConn).ReadString('\n')
		if err != nil {
			fmt.Println(err)
		}
		command := strings.Trim(status, " \n")
		fmt.Printf("Command from master: %v\n", command)
		switch command {
		case "get":
			fmt.Println("Processing 'get' from master")

		case "alive":
			fmt.Println("Processing 'alive' from master")
			doAlive()
			fmt.Println("Done responding to master")

		case "broadcast":
			fmt.Println("Processing 'broadcast' from master")
		}
		// get
		// alive
		// broadcast

		// 		// run loop forever (or until ctrl-c)
		// 		for {
		// 			// will listen for message to process ending in newline (\n)
		// 			message, _ := bufio.NewReader(conn).ReadString('\n')
		// 			// output message received
		// 			fmt.Print("Message Received:", string(message))
		// 			// sample process for string received
		// 			newmessage := strings.ToUpper(message)
		// 			// send new string back to client
		// 			conn.Write([]byte(newmessage + "\n"))
		// 		}
	}
}
