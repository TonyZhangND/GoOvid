package main

import (
	"os"
	"strconv"
	"strings"

	comm "github.com/TonyZhangND/GoOvid/commons"
	conf "github.com/TonyZhangND/GoOvid/configs"
	serv "github.com/TonyZhangND/GoOvid/server"
)

func main() {
	// process command line arguments and parse config
	config := strings.Trim(os.Args[1], " ")
	myBox := comm.ParseBoxAddr(strings.Trim(os.Args[2], " "))
	masterPort, err := strconv.ParseUint(os.Args[3], 10, 16)
	comm.CheckFatalOvidErrorf(err, "Cannot parse masterPort %v (%v)\n", os.Args[3], err)
	agentMap := *conf.Parse(config)

	// get a list of known boxes
	knownBoxSet := make(map[comm.BoxID]int)
	for _, agent := range agentMap {
		knownBoxSet[agent.Box] = 1
	}
	knownBoxes := make([]comm.BoxID, len(knownBoxSet))
	i := 0
	for bid := range knownBoxSet {
		knownBoxes[i] = bid
		i++
	}
	serv.InitAndRunServer(myBox, knownBoxes, comm.PortNum(masterPort))
}
