package main

import (
	"flag"

	comm "github.com/TonyZhangND/GoOvid/commons"
	conf "github.com/TonyZhangND/GoOvid/configs"
	serv "github.com/TonyZhangND/GoOvid/server"
)

func main() {
	// process command line arguments and parse config
	masterPort := flag.Int("master", 0, "Local port number for master connection")
	flag.Parse()
	config := flag.Args()[0]
	myBox := comm.ParseBoxAddr(flag.Args()[1])

	mp := comm.PortNum(*masterPort)
	agentMap := *conf.Parse(config)

	// start only if box is valid
	for _, agent := range agentMap {
		if agent.Box == myBox {
			serv.InitAndRunServer(myBox, agentMap, mp)
			return
		}
	}
	comm.FatalOvidErrorf("Box %v not in configuration %s\n", myBox, config)
}
