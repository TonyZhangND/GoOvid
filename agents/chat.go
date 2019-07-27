package agents

import (
	"bufio"
	"fmt"
	"os"

	"github.com/TonyZhangND/GoOvid/server"
)

type chatAgent struct {
	broadcast func(msg string)
	userName  string
}

func newChatAgent(name string, broadcastFunc func(msg string)) *chatAgent {
	return &chatAgent{broadcast: broadcastFunc, userName: name}
}

func (ca *chatAgent) deliver(msg string) {}

func (ca *chatAgent) run() {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("%v > ", ca.userName)
		// Read the keyboad input.
		input, err := reader.ReadString('\n')
		if err != nil {
			server.FatalError(fmt.Sprintf("Invalid input %v in chatAgent", input))
		}
	}
}
