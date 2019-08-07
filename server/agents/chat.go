package agents

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	c "github.com/TonyZhangND/GoOvid/commons"
)

// ChatAgent struct
type ChatAgent struct {
	send             func(vDest c.ProcessID, msg string)
	fatalAgentErrorf func(errMsg string, a ...interface{})
	userName         string
	contacts         []c.ProcessID
	isActive         bool
}

// Init fills the empty struct with this agent's fields and attributes.
func (ca *ChatAgent) Init(attrs map[string]interface{},
	send func(vDest c.ProcessID, msg string),
	fatalAgentErrorf func(errMsg string, a ...interface{})) {
	ca.send = send
	ca.fatalAgentErrorf = fatalAgentErrorf
	ca.userName = attrs["myname"].(string)
	ca.contacts = make([]c.ProcessID, len(attrs["contacts"].([]interface{})))
	for i, id := range attrs["contacts"].([]interface{}) {
		ca.contacts[i] = c.ProcessID(id.(float64))
	}
	ca.isActive = false
}

// Halt stops the execution of ca.
func (ca *ChatAgent) Halt() {
	ca.isActive = false
}

// Deliver a message of the format "<sender name> <contents>".
// The chat agent ignores the port.
func (ca *ChatAgent) Deliver(data string, port c.PortNum) {
	dataSlice := strings.SplitN(strings.TrimSpace(data), " ", 2)
	sender, msg := dataSlice[0], dataSlice[1]
	fmt.Printf("\n%s > %s\n%s > ", sender, msg, ca.userName)
}

// Run begins the execution of the ca agent.
func (ca *ChatAgent) Run() {
	reader := bufio.NewReader(os.Stdin)
	ca.isActive = true
	for ca.isActive {
		fmt.Printf("%v > ", ca.userName)
		// Read the keyboad input.
		input, err := reader.ReadString('\n')
		if err != nil {
			ca.fatalAgentErrorf("Invalid input %v in chatAgent\n", input)
		}
		for _, vDest := range ca.contacts {
			ca.send(vDest, fmt.Sprintf("%s %s", ca.userName, strings.TrimSpace(input)))
		}
	}
}
