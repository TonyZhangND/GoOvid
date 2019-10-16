package main

import (
	"flag"
	"fmt"

	agnt "github.com/TonyZhangND/GoOvid/agents"
	comm "github.com/TonyZhangND/GoOvid/commons"
	conf "github.com/TonyZhangND/GoOvid/configs"
	serv "github.com/TonyZhangND/GoOvid/server"
)

// Prints the resulting map from running Parse()
func printResult(res map[comm.ProcessID]agnt.AgentInfo) {
	fmt.Println("\nPrinting agent map:")
	for k, v := range res {
		fmt.Printf("%v : %v\n", k, v)
	}
	fmt.Println("")
}

func main() {
	// process command line arguments and parse config
	masterPort := flag.Int("master", 0, "Local port number for master connection")
	debugMode := flag.Bool("debug", false, "Toggles debugMode to on")
	loss := flag.Float64("loss", 0, "Rate at which a server drops inter-agent messages")
	flag.Parse()
	config := flag.Args()[0]
	myBox := comm.ParseBoxAddr(flag.Args()[1])

	mp := comm.PortNum(*masterPort)
	if *loss < 0 || *loss > 1. {
		comm.FatalOvidErrorf("loss must be in range 0-1\n")
	}
	agentMap := conf.Parse(config)
	// printResult(agentMap)

	// start only if box is valid
	for _, agent := range agentMap {
		if agent.Box == myBox {
			if *debugMode {
				serv.DebugMode = true
			}
			serv.InitAndRunServer(myBox, agentMap, mp, *loss)
			return
		}
	}
	comm.FatalOvidErrorf("Box %v not in configuration %s\n", myBox, config)
}
