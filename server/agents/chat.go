package agents

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	c "github.com/TonyZhangND/GoOvid/commons"
)

type chatAgent struct {
	send             func(dest c.ProcessID, msg string)
	fatalAgentErrorf func(errMsg string, a ...interface{})
	userName         string
	contacts         []c.ProcessID
	isActive         bool
}

// Init fills the empty struct with this agent's fields and attributes
func (ca *chatAgent) Init(attrs map[string]interface{},
	send func(dest c.ProcessID, msg string),
	fatalAgentErrorf func(errMsg string, a ...interface{})) {
	ca.fatalAgentErrorf = fatalAgentErrorf
	ca.userName = attrs["myname"].(string)
	ca.contacts = make([]c.ProcessID, len(attrs["contacts"].([]interface{})))
	for i, id := range attrs["contacts"].([]interface{}) {
		ca.contacts[i] = c.ProcessID(id.(int))
	}
	ca.isActive = false
}

// Halt stops the execution of ca
func (ca *chatAgent) Halt() {
	ca.isActive = false
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
			ca.fatalAgentErrorf("Invalid input %v in chatAgent\n", input)
		}
		for _, agent := range ca.contacts {
			ca.send(agent, fmt.Sprintf("%s %s", ca.userName, input))
		}
	}
}
