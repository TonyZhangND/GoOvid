package paxos

// This file describes the states and transitions of a paxos replica that is related
// to its role as an acceptor

import (
	"fmt"

	c "github.com/TonyZhangND/GoOvid/commons"
)

type leaderState struct {
	ballotNum *ballot
	active    bool
	proposals map[string]*proposal // given k->*v, k is a hash of v
	propose   chan proposal
	adopted   chan adoptedPayload
	preempted chan ballot
	p1bChan   chan string // channel to push p1b messages to scout
}

type adoptedPayload struct {
	ballotNum ballot
	pvals     []pValue
}

// Constructor
func (rep *ReplicaAgent) newLeaderState() *leaderState {
	return &leaderState{
		ballotNum: &ballot{rep.myID, 0},
		active:    false,
		proposals: make(map[string]*proposal),
		propose:   make(chan proposal),
		adopted:   make(chan adoptedPayload),
		preempted: make(chan ballot)}
}

func (rep *ReplicaAgent) spawnScout(baln uint64, p1bChan chan string) {
	waitfor := make(map[c.ProcessID]bool) // set of acceptors from which p1b is pending
	for acc := range rep.replicas {
		// Send "p1a <sender> <balNum>"
		waitfor[acc] = true
		p1a := fmt.Sprintf("p1a %d %d", rep.myID, baln)
		rep.send(acc, p1a)
	}
	// for rep.isActive {
	// 	payload :=
	// }
}
