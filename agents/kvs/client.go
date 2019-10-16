package kvs

// This file contains the definition and logic of a client agent.
// A client agent acts as the client of a key value store. It sends requests to the kvs,
// and processes the responses.

import (
	"fmt"
	"strings"

	c "github.com/TonyZhangND/GoOvid/commons"
)

// ClientAgent struct contains the information inherent to a kvs agent
type ClientAgent struct {
	send             func(vDest c.ProcessID, msg string)
	fatalAgentErrorf func(errMsg string, a ...interface{})
	debugPrintf      func(s string, a ...interface{})
	isActive         bool
	myID             c.ProcessID
}

// Init fills the empty clt struct with this agent's fields and attributes.
func (clt *ClientAgent) Init(attrs map[string]interface{},
	send func(vDest c.ProcessID, msg string),
	fatalAgentErrorf func(errMsg string, a ...interface{}),
	debugPrintf func(s string, a ...interface{})) {
	clt.send = send
	clt.fatalAgentErrorf = fatalAgentErrorf
	clt.debugPrintf = debugPrintf
	clt.isActive = false
	clt.myID = c.ProcessID(attrs["myid"].(float64))
}

// Halt stops the execution of clt.
func (clt *ClientAgent) Halt() {
	clt.isActive = false
}

// Deliver a message from either a tty or kvs agent
// The client agent expects :
//   - tty agent at virtual dest 1
//   - kvs agent at virtual dest 2
//   - tty commands to enter via port 1
//		> commands are of the format "put <key> <value>" or "get <key>", where
//		  <key> does not contain spaces
//   - kvs responses to enter via port 2
//   	> replies are of the format "putRes" which represents a sucessful put, or
//        getRes <val> where <val> is the result of the getRes.
func (clt *ClientAgent) Deliver(request string, port c.PortNum) {
	clt.debugPrintf("Client received request %s\n", request)
	switch port {
	case 1: //tty command -> forward to kvs
		msg := fmt.Sprintf("%v %s", clt.myID, request)
		clt.send(2, msg)
	case 2: //kvs response -> forward to tty
		repSlice := strings.SplitN(request, " ", 2)
		switch repSlice[0] {
		case "putok": //successful put
			clt.send(1, "ok")
		case "getok":
			clt.send(1, repSlice[1])
		case "getbad":
			clt.send(1, "No such entry")
		default:
			clt.Halt()
			clt.fatalAgentErrorf("Unexpected response %s from kvs\n", request)
		}
	default:
		clt.Halt()
		clt.fatalAgentErrorf("Unexpected message %s in port %v\n", request, port)
	}
}

// Run begins the execution of the clt agent.
func (clt *ClientAgent) Run() {
	clt.isActive = false
}
