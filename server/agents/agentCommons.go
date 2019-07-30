package agents

import (
	"fmt"
	"net"
	"os"
	"runtime/debug"

	c "github.com/TonyZhangND/GoOvid/commons"
)

const debugMode = true

type AgentType int

const (
	Chat AgentType = 1
)

type Agent interface {
	deliver(msg string)
	run()
	halt()
	name() string
}

type AgentInfo struct {
	agentType AgentType
	hostIP    net.IP
	hostPort  c.PortNum
}

// DebugPrintln prints the string s if debug mode is on
func debugPrintln(agent *Agent, s string) {
	if debugMode {
		fmt.Printf("Agent %v : %v\n", (*agent).name())
	}
}

// fatalError prints the error messange and halts the agent
func fatalError(agent Agent, errMsg string) {
	agent.halt()
	fmt.Printf("Error : Agent %v : %v\n", agent.name(), errMsg)
	debug.PrintStack()
	os.Exit(1)
}