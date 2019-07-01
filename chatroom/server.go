package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

type server struct {
	physID     uint16
	gridSize   uint16
	masterIP   string
	masterPort uint16
	gridIP     string
	shouldRun  bool
}

// newServer is the constructor for server.
// It returns a server struct with default values for some fields.
func newServer(physID uint16, gridSize uint16, masterPort uint16) server {
	return server{
		uint16(physID),
		uint16(gridSize),
		"127.0.0.1",
		uint16(masterPort),
		"127.0.0.1",
		true}
}

// String is the "toString" method for a server object
// It returns a string describing this server s.
func (s server) String() string {
	return fmt.Sprintf("* GoOvid server *\n"+
		"physID: %d\n"+
		"gridSize: %d\n"+
		"masterPort: %d\n",
		s.physID, s.gridSize, s.masterPort)
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
	server := newServer(
		uint16(pid),
		uint16(gridSize),
		uint16(masterPort))
	fmt.Println(server)

	// listen for master on the master address
	masterAddr := fmt.Sprintf("%s:%d", server.masterIP, server.masterPort)
	fmt.Println("Listening for master connecting on " + masterAddr)
	masterListener, _ := net.Listen("tcp", masterAddr)
	masterConn, _ := masterListener.Accept()
	defer masterConn.Close()
	fmt.Println("Accepted master connection")

	for server.shouldRun {
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
