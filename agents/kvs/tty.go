package kvs

// This file contains the definition and logic of a tty agent.
// A tty agent is used to interact with the client of a key value store.
// It forwards any input to the client, and prints any messages received
// from the client. The agent blocks user inputs until a response is received
// from each new command

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	c "github.com/TonyZhangND/GoOvid/commons"
)

// TTYAgent struct contains the information inherent to a tty agent
type TTYAgent struct {
	send             func(vDest c.ProcessID, msg string)
	fatalAgentErrorf func(errMsg string, a ...interface{})
	debugPrintf      func(s string, a ...interface{})
	isActive         bool
	block            bool
}

// Init fills the empty tty struct with this agent's fields and attributes.
func (tty *TTYAgent) Init(attrs map[string]interface{},
	send func(vDest c.ProcessID, msg string),
	fatalAgentErrorf func(errMsg string, a ...interface{}),
	debugPrintf func(s string, a ...interface{})) {
	tty.send = send
	tty.fatalAgentErrorf = fatalAgentErrorf
	tty.isActive = false
	tty.block = false
}

// Halt stops the execution of tty.
func (tty *TTYAgent) Halt() {
	tty.isActive = false
}

// Deliver and print a message from the client agent .
// The tty agent expects:
// - client agent at virtual dest 1
// - client responses to enter via port 1
func (tty *TTYAgent) Deliver(data string, port c.PortNum) {
	fmt.Printf("%s\n", data)
	tty.block = false
}

// Run begins the execution of the tty agent.
// commands are of the format "put <key> <value>" or "get <key>", where
//		  <key> does not contain spaces
func (tty *TTYAgent) Run() {
	reader := bufio.NewReader(os.Stdin)
	tty.isActive = true
	for tty.isActive {
		for tty.block {
			time.Sleep(500 * time.Millisecond)
		}
		fmt.Printf("> ")
		// Read the keyboad input and check for validity
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Invalid command\n")
			continue
		}
		reqSlice := strings.SplitN(strings.TrimSpace(input), " ", 3)
		var req string
		switch reqSlice[0] {
		case "put":
			if len(reqSlice) != 3 {
				fmt.Printf("Invalid command\n")
				continue
			}
			req = fmt.Sprintf("put %s %s", reqSlice[1], reqSlice[2])
		case "get":
			if len(reqSlice) != 2 {
				fmt.Printf("Invalid command\n")
				continue
			}
			req = fmt.Sprintf("get %s", reqSlice[1])
		default:
			fmt.Printf("Invalid command\n")
			continue
		}
		// Command is valid. Send to client and wait for response
		tty.block = true
		tty.send(1, req)
	}
}
