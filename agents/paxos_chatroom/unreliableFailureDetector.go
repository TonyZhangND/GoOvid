package paxos

import (
	"fmt"
	"time"

	c "github.com/TonyZhangND/GoOvid/commons"
)

type pingTimer struct {
	inChan chan bool
}

type unreliableFailureDetector struct {
	replica *ReplicaAgent              // agent this ufd is bound to
	alive   map[c.ProcessID]*pingTimer // alive[q]= pt iff q is thought to be alive
	leaders map[c.ProcessID]bool       // set of processes believed to be the leader
}

// Constructor for a new unreliableFailureDetector
func newUnreliableFailureDetector(rep *ReplicaAgent) *unreliableFailureDetector {
	ufd := unreliableFailureDetector{}
	ufd.alive = make(map[c.ProcessID]*pingTimer)
	ufd.leaders = make(map[c.ProcessID]bool)
	ufd.replica = rep
	return &ufd
}

// Begins sending pings into l.conn channel
func (ufd *unreliableFailureDetector) runPinger() {
	for ufd.replica.isActive {
		var ping string
		if _, ok := ufd.leaders[ufd.replica.myID]; ok {
			// Replica believes that it is the leader
			ping = fmt.Sprintf("ping %d leader", ufd.replica.myID)
		} else {
			ping = fmt.Sprintf("ping %d", ufd.replica.myID)
		}
		// Broadcast ping
		for rep := range ufd.replica.replicas {
			if rep != ufd.replica.myID {
				ufd.replica.send(rep, ping)
			}
		}
		time.Sleep(pingInterval)
	}
}

// TODO: handle a ping of format "ping <sender> [leader]"
func (ufd *unreliableFailureDetector) receivePing(ping string) {
	// Idea is to push into pingTimer channel if process is thought to be alive,
	// else add process p into alive set with alive[p] = new ping timer
}
