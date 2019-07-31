package agents

import (
	"bufio"
	"fmt"
	"os"
)

type chatAgent struct {
	broadcast func(msg string)
	userName  string
	isActive  bool
}

func (ca *chatAgent) agentInit(attrs map[string]interface{}, broadcast func(msg string),
	send func(msg string, dest AgentType)) {
}
func (ca *chatAgent) halt()              {}
func (ca *chatAgent) name() string       { return "" }
func (ca *chatAgent) deliver(msg string) {}

func (ca *chatAgent) run() {
	reader := bufio.NewReader(os.Stdin)
	ca.isActive = true
	for ca.isActive {
		fmt.Printf("%v > ", ca.userName)
		// Read the keyboad input.
		input, err := reader.ReadString('\n')
		if err != nil {
			fatalAgentErrorf(ca, "Invalid input %v in chatAgent\n", input)
		}
	}
}
