package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"

	c "github.com/TonyZhangND/GoOvid/commons"
	a "github.com/TonyZhangND/GoOvid/server/agents"
)

func checkDecodeError(err error, dat string) {
	c.CheckFatalOvidError(err, fmt.Sprintf("%v encountered decoding %v", err, dat))
}

func Parse(configFile string) *map[c.ProcessID]*a.AgentInfo {
	// Read the file
	dat, err := ioutil.ReadFile(configFile)
	if err != nil {
		c.FatalOvidError(fmt.Sprintf("Error: %v encountered reading %v", err, configFile))
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
		agent := a.AgentInfo{} // alloc empty struct for pid

		// loop over the json object for pid
		objM := obj.(map[string]interface{})
		for k, v := range objM {
			switch k {
			case "type":
				switch v.(string) {
				case "chat":
					agent.AgentType = a.Chat
				default:
					c.FatalOvidError(fmt.Sprintf("Unknown agent type %v", v))
				}
			case "box":
				agent.Box = v.(string)
			case "attrs":
				agent.RawAttrs = v.(map[string]interface{})
			case "routes":
				agent.Routes = v.(map[string]interface{})
			default:
				c.FatalOvidError(fmt.Sprintf("Unknown agent field %v", k))
			}
		}
		res[c.ProcessID(pid)] = &agent
	}
	fmt.Println(res)
	for k, v := range res {
		fmt.Printf("%v : %v\n", k, v)
	}
	return &res
}

func main() {
	Parse("chat.json")
}
