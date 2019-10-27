package paxos

// This file contains the definition and logic of a centralized controller of the
// paxos service.
// The ControllerAgent type must implement the Agent interface.

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	c "github.com/TonyZhangND/GoOvid/commons"
)

// ControllerAgent struct contains the information inherent to a controller
type ControllerAgent struct {
	send             func(vDest c.ProcessID, msg string)
	fatalAgentErrorf func(errMsg string, a ...interface{})
	debugPrintf      func(s string, a ...interface{})
	isActive         bool
	clients          map[c.ProcessID]int // using map because we want search capability
	replicas         map[c.ProcessID]int
}

// Init fills the empty ctr struct with this agent's fields and attributes.
func (ctr *ControllerAgent) Init(attrs map[string]interface{},
	send func(vDest c.ProcessID, msg string),
	fatalAgentErrorf func(errMsg string, a ...interface{}),
	debugPrintf func(s string, a ...interface{})) {
	ctr.send = send
	ctr.fatalAgentErrorf = fatalAgentErrorf
	ctr.debugPrintf = debugPrintf
	ctr.isActive = false

	// Parse and store attributes
	ctr.clients, ctr.replicas = make(map[c.ProcessID]int), make(map[c.ProcessID]int)
	for _, x := range attrs["clients"].([]interface{}) {
		id := c.ProcessID(x.(float64))
		ctr.clients[id] = 0
	}
	for _, x := range attrs["replicas"].([]interface{}) {
		id := c.ProcessID(x.(float64))
		ctr.replicas[id] = 0
	}
}

// Halt stops the execution of the agent.
func (ctr *ControllerAgent) Halt() {
	ctr.isActive = false
}

// Deliver a message
func (ctr *ControllerAgent) Deliver(request string, port c.PortNum) {

}

// Run begins the execution of the paxos agent.
func (ctr *ControllerAgent) Run() {
	reader := bufio.NewReader(os.Stdin)
	ctr.isActive = true
	fmt.Println("Paxos controller active. Enter your command")
	for ctr.isActive {
		fmt.Printf("> ")
		// Read the keyboad input.
		input, err := reader.ReadString('\n')
		if err != nil {
			ctr.Halt()
			ctr.fatalAgentErrorf("Invalid input %v in controller\n", input)
			os.Exit(0)
		}
		if len(strings.TrimSpace(input)) < 1 {
			// Ignore empty messages
			continue
		}
		inputSlice := strings.SplitN(strings.TrimSpace(input), " ", 2)
		command := inputSlice[0]
		switch command {
		case "exit":
			fmt.Println("Terminating paxos cluster")
			ctr.Halt()
			os.Exit(0)
		case "start":
			// Start a node
			if len(inputSlice) < 2 {
				fmt.Println("Invalid input")
				continue
			}
			nodePort, err := strconv.ParseUint(inputSlice[1], 10, 64)
			if err != nil {
				fmt.Println("Invalid input")
				continue
			}
			box := fmt.Sprintf("127.0.0.1:%d", nodePort)
			proc := exec.Command("./ovid", "-log", "configs/paxos.json", box)
			proc.Stdout = os.Stdout
			err = proc.Start()
			if err != nil {
				fmt.Printf("Failed to start 127.0.0.1:%d : %v\n", nodePort, err)
				continue
			}
			fmt.Printf("Started box 127.0.0.1:%d\n", nodePort)
		case "req":
			// Issue a client request
			if len(inputSlice) < 2 {
				fmt.Println("Invalid input")
				continue
			}
			payload := strings.SplitN(inputSlice[1], " ", 2)
			if len(payload) < 2 {
				fmt.Println("Invalid input")
				continue
			}
			destUint, err := strconv.ParseUint(payload[0], 10, 64)
			if err != nil {
				fmt.Printf("Invalid request destination %v\n", payload[0])
			}
			dest := c.ProcessID(destUint)
			_, ok := ctr.clients[dest]
			if !ok {
				fmt.Printf("Invalid client %v\n", dest)
				continue
			}
			m := payload[1]
			ctr.send(dest, fmt.Sprintf("issue %s", m))
		case "kill": // Issue a kill command
			payload := strings.Split(inputSlice[1], " ")
			if len(payload) > 1 {
				fmt.Println("Invalid input")
				continue
			}
			target := payload[0]
			if target == "primary" {
				for rep := range ctr.replicas {
					ctr.send(rep, "kill primary")
				}
			} else { // assasinate a specific target
				tango, err := strconv.ParseUint(target, 10, 64)
				if err != nil {
					fmt.Printf("Invalid assasination target %v\n", target)
					continue
				}
				_, ok := ctr.replicas[c.ProcessID(tango)]
				if !ok {
					fmt.Printf("Invalid assasination target %v\n", target)
					continue
				}
				ctr.send(c.ProcessID(tango), fmt.Sprintf("kill"))
			}
		case "dump":
			if len(inputSlice) > 1 {
				fmt.Println("Invalid input")
				continue
			}
			for rep := range ctr.replicas {
				ctr.send(rep, fmt.Sprintf("dump"))
			}
		case "skip":
			if len(inputSlice) < 2 {
				fmt.Println("Invalid input")
				continue
			}
			payload := strings.SplitN(inputSlice[1], " ", 2)
			if len(payload) < 2 {
				fmt.Println("Invalid input")
				continue
			}
			destUint, err := strconv.ParseUint(payload[0], 10, 64)
			if err != nil {
				fmt.Printf("Invalid request destination %v\n", payload[0])
			}
			dest := c.ProcessID(destUint)
			_, ok := ctr.replicas[dest]
			if !ok {
				fmt.Printf("Invalid replica %v\n", dest)
				continue
			}
			slot, err := strconv.ParseUint(payload[1], 10, 64)
			if err != nil {
				fmt.Printf("Invalid slot %v\n", inputSlice[1])
				continue
			}
			ctr.send(dest, fmt.Sprintf("skip %d", slot))
		default:
			fmt.Println("Invalid command")
			continue
		}
	}
}
