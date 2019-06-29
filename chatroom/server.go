package main

import (
	"fmt"
	"os"
	"strconv"
)

// only needed below for sample processing

type server struct {
	physID     uint16
	gridSize   uint16
	masterPort uint16
}

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
		errStr := fmt.Sprintf("Errors occured while processing arguments.\n"+
			"PhysID: %v\n"+
			"gridSize: %v\n"+
			"masterPort: %v\n"+
			"Program exiting...",
			err1, err2, err3)
		fmt.Println(errStr)
		os.Exit(1)
	}

	server := server{uint16(pid), uint16(gridSize), uint16(masterPort)}
	fmt.Println("Launching server...")
	fmt.Println(server)

	// // listen on all interfaces
	// ln, _ := net.Listen("tcp", "127.0.0.1:8081")

	// // accept connection on port
	// conn, _ := ln.Accept()

	// // run loop forever (or until ctrl-c)
	// for {
	// 	// will listen for message to process ending in newline (\n)
	// 	message, _ := bufio.NewReader(conn).ReadString('\n')
	// 	// output message received
	// 	fmt.Print("Message Received:", string(message))
	// 	// sample process for string received
	// 	newmessage := strings.ToUpper(message)
	// 	// send new string back to client
	// 	conn.Write([]byte(newmessage + "\n"))
	// }
}
