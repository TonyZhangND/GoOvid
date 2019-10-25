package paxos

// This file describes the states and transitions of a paxos replica that is related
// to its role as an acceptor

import (
	"fmt"
	"math"

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

// Payload for "adopted" messages from scout to leader
type adoptedPayload struct {
	ballotNum ballot
	pvals     map[uint64]pValue
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

// Scout process in Fig 6 of PMMC
func (rep *ReplicaAgent) spawnScout(baln uint64, p1bChan chan string) {
	waitfor := make(map[c.ProcessID]bool) // set of acceptors from which p1b is pending
	myBallot := &ballot{rep.myID, baln}
	processedPVals := make(map[uint64]pValue)
	for acc := range rep.replicas {
		// Send "p1a <sender> <balNum>"
		waitfor[acc] = true
		p1a := fmt.Sprintf("p1a %d %d", myBallot.id, myBallot.n)
		rep.send(acc, p1a)
	}
	for rep.isActive {
		payload := <-p1bChan
		acc, ballot, pVals := parseP1bPayload(payload)
		if myBallot.id == ballot.id && myBallot.n == ballot.n {
			// Adopted :) Now merge pValues from acceptor. For each p in pVals
			// 1. If p.slot not in rprocessedPVals then processedPVals[p.slot] = p
			// 2. Else, if processedPVals[p.slot].ballot.lt(p.ballot) then
			//    processedPVals[p.slot] = p
			for s, p := range pVals {
				if rpv, ok := processedPVals[s]; !ok {
					processedPVals[s] = *p
				} else {
					if rpv.ballot.lt(p.ballot) {
						processedPVals[s] = *p
					}
				}
			}
			// Mark acc as responded
			delete(waitfor, acc)
			if len(waitfor) <= int(math.Floor(float64(len(rep.replicas))/2.0)) {
				a := adoptedPayload{*myBallot, processedPVals}
				rep.leader.adopted <- a
				return
			}
		} else {
			// Pre-empted :(
			rep.leader.preempted <- *ballot
			return
		}
	}
}
