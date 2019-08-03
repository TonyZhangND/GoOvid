package agents

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type chatAgent struct {
	broadcast func(msg string)
	userName  string
	isActive  bool
}

// Init fills the empty struct with this agent's fields and attributes
func (ca *chatAgent) Init(attrs map[string]interface{}, broadcast func(msg string)) {
	ca.userName = attrs["myname"].(string)
	ca.broadcast = broadcast
	ca.isActive = false
}

// Halt stops the execution of ca
func (ca *chatAgent) Halt() {
	ca.isActive = false
}

// Name returns the username of ca
func (ca *chatAgent) Name() string {
	return ca.userName
}

// Deliver a message of the format "<sender name> <contents>"
func (ca *chatAgent) Deliver(data string) {
	dataSlice := strings.SplitN(strings.TrimSpace(data), " ", 2)
	sender, msg := dataSlice[0], dataSlice[1]
	fmt.Printf("%s > %s\n", sender, msg)
}

// Run begins the execution of the ca agent
func (ca *chatAgent) Run() {
	reader := bufio.NewReader(os.Stdin)
	ca.isActive = true
	for ca.isActive {
		fmt.Printf("%v > ", ca.userName)
		// Read the keyboad input.
		input, err := reader.ReadString('\n')
		if err != nil {
			fatalAgentErrorf(ca, "Invalid input %v in chatAgent\n", input)
		}
		ca.broadcast(input)
	}
}
