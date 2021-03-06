package kvs

// This file contains the definition and logic of a kvs agent.
// A kvs implements a key-value-store with a "put" and "get" API. It ensures data
// durability by maintaining an append-only log.
// The KVSAgent type must implement the Agent interface.
// Requirement: keys do not contain whitespace

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	c "github.com/TonyZhangND/GoOvid/commons"
)

// ReplicaAgent struct contains the information inherent to a kvs replica
type ReplicaAgent struct {
	send             func(vDest c.ProcessID, msg string)
	fatalAgentErrorf func(errMsg string, a ...interface{})
	debugPrintf      func(s string, a ...interface{})
	isActive         bool
	inMemoryStore    map[string]string
	logFile          *os.File //log file
	logger           *log.Logger
}

// Init fills the empty kvs struct with this agent's fields and attributes.
func (kvs *ReplicaAgent) Init(attrs map[string]interface{},
	send func(vDest c.ProcessID, msg string),
	fatalAgentErrorf func(errMsg string, a ...interface{}),
	debugPrintf func(s string, a ...interface{})) {
	kvs.send = send
	kvs.fatalAgentErrorf = fatalAgentErrorf
	kvs.debugPrintf = debugPrintf
	kvs.isActive = false
	kvs.inMemoryStore = make(map[string]string)
	logPath := attrs["log"].(string)
	// TODO: if existing files exist, build store from it
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		kvs.Halt()
		kvs.fatalAgentErrorf("Cannot create file %v\n", logPath)
	}
	kvs.logFile = logFile
	kvs.logger = log.New(kvs.logFile, "", log.LstdFlags)
}

// Halt stops the execution of kvs.
func (kvs *ReplicaAgent) Halt() {
	kvs.isActive = false
	kvs.logFile.Close()
}

// Deliver a message of the format "<sender physical id> get <key>" or
// "<sender physical id> put <key> <data>"
// The kvs agent expects all client requests to enter via port 1.
func (kvs *ReplicaAgent) Deliver(request string, port c.PortNum) {
	kvs.debugPrintf("KVS received request %s\n", request)
	if port != 1 {
		kvs.fatalAgentErrorf("Unexpected request %s in port %v\n", request, port)
	}
	reqSlice := strings.SplitN(strings.TrimSpace(request), " ", 3)
	senderStr, requestType, data := reqSlice[0], reqSlice[1], reqSlice[2]
	sender, err := strconv.Atoi(senderStr)
	if err != nil {
		kvs.Halt()
		kvs.fatalAgentErrorf("Cannot parse sender in request %s\n", request)
	}
	switch requestType {
	case "put":
		dataSlice := strings.SplitN(data, " ", 2)
		key, val := dataSlice[0], dataSlice[1]
		// Store and ppend data to log
		kvs.inMemoryStore[key] = val
		kvs.logger.Printf("%v %s %s\n", sender, key, val)
		// Reply to client
		kvs.send(c.ProcessID(sender), "putok")
	case "get":
		key := strings.TrimSpace(data)
		val, ok := kvs.inMemoryStore[key]
		if !ok {
			// No value for such a key
			kvs.send(c.ProcessID(sender), "getbad")
		} else {
			// Key exists
			reply := fmt.Sprintf("getok %s", val)
			kvs.send(c.ProcessID(sender), reply)
		}
	default:
		kvs.Halt()
		kvs.fatalAgentErrorf("Invalid request %v\n", request)
	}
}

// Run begins the execution of the kvs agent.
func (kvs *ReplicaAgent) Run() {
	kvs.isActive = true
}
