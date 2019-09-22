package configs

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"

	a "github.com/TonyZhangND/GoOvid/agents"
	c "github.com/TonyZhangND/GoOvid/commons"
)

// Prints the error messange and kills the program
// if an error is detected
func checkDecodeError(err error, dat string) {
	c.CheckFatalOvidErrorf(err, "%v encountered decoding %v", err, dat)
}

// Helper: Parses the json object of an agent, returning a pointer to the
// resulting AgentInfo struct
func parseAgentObject(agentObj map[string]interface{}) *a.AgentInfo {
	agent := &a.AgentInfo{} // alloc empty struct for the agent
	for k, v := range agentObj {
		switch k {
		case "type":
			switch v.(string) {
			case "chat":
				agent.Type = a.Chat
			case "dummy":
				agent.Type = a.Dummy
			case "kvs":
				agent.Type = a.KVS
			case "client":
				agent.Type = a.Client
			case "tty":
				agent.Type = a.TTY
			default:
				c.FatalOvidErrorf("Unknown agent type %v\n", v)
			}
		case "box":
			agent.Box = c.ParseBoxAddr(v.(string))
		case "attrs":
			agent.RawAttrs = v.(map[string]interface{})
		case "routes":
			// initialize the routing table
			routingTable := make(map[c.ProcessID]c.Route)

			// iterate over each link
			rts := v.(map[string]interface{})
			for vidRaw, rtRaw := range rts {
				vid, err := strconv.ParseUint(vidRaw, 10, 16)
				checkDecodeError(err, vidRaw)
				rt := rtRaw.(map[string]interface{})
				if len(rt) != 1 {
					c.FatalOvidErrorf("Invalid route entry %v\n", rtRaw)
				}
				// parse the json object for the link
				route := c.Route{} // alloc a Route struct to be filled
				for pidRaw, portRaw := range rt {
					pid, err := strconv.ParseUint(pidRaw, 10, 16)
					checkDecodeError(err, pidRaw)
					port := c.PortNum(portRaw.(float64))
					route.DestID = c.ProcessID(pid)
					route.DestPort = c.PortNum(port)
				}
				routingTable[c.ProcessID(vid)] = route
			}
			agent.Routes = routingTable
		default:
			c.FatalOvidErrorf("Unknown agent field %v\n", k)
		}
	}
	return agent
}

// IsValid returns false if config is detected as invalid. Otherwise returns true.
func isValid(config map[c.ProcessID]*a.AgentInfo) (bool, error) {

	// Check for routes pointing to non-existent agents
	for pid, agent := range config {
		for vdest, route := range agent.Routes {
			pdest := route.DestID
			if _, ok := config[pdest]; !ok {
				msg := fmt.Sprintf("Invalid destination %v : { %v : %v } in routing table of agent %v",
					vdest, pdest, route.DestPort, pid)
				return false, errors.New(msg)
			}
		}
	}
	return true, nil
}

// Parse reads the ovid configuration in configFile, and returns a pointer
// to a map containing the AgentInfo objects in the configuration
func Parse(configFile string) map[c.ProcessID]*a.AgentInfo {
	// Read the file
	dat, err := ioutil.ReadFile(configFile)
	if err != nil {
		c.FatalOvidErrorf("%v \n", err)
	}

	// Decode the file into a map[string]interface{}
	var rawMap interface{}
	err = json.Unmarshal(dat, &rawMap)
	checkDecodeError(err, configFile)

	// Decode map[string]interface{} into a new AgentInfo struct, and return a
	// map containing all the agents
	res := make(map[c.ProcessID]*a.AgentInfo)
	m := rawMap.(map[string]interface{})
	for id, obj := range m {
		pid, err := strconv.ParseUint(id, 10, 16)
		checkDecodeError(err, configFile)
		res[c.ProcessID(pid)] = parseAgentObject(obj.(map[string]interface{}))
	}

	// check for validity
	ok, err := isValid(res)
	if !ok {
		c.FatalOvidErrorf("%v \n", err)
	}
	return res
}
