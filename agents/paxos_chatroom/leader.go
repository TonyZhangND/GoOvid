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
		propose:   make(chan proposal)}
}

// Scout process in Fig 6 of PMMC
func (rep *ReplicaAgent) spawnScout(
	baln uint64,
	preempted chan ballot,
	adopted chan adoptedPayload,
	p1bChan chan string) {

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
		if myBallot.eq(ballot) {
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
				adopted <- a
				return
			}
		} else {
			// Pre-empted :(
			preempted <- *ballot
			return
		}
	}
}

// Commander process in Fig 6 of PMMC
func (rep *ReplicaAgent) spawnCommander(
	pval *pValue, preempted chan ballot,
	p2bChan chan string) {

	waitfor := make(map[c.ProcessID]bool) // set of acceptors from which p2b is pending
	myBallot := pval.ballot

	for acc := range rep.replicas {
		// Send "p2a <balID> <balNum> <slot> <clientID> <reqNum> <m>"
		waitfor[acc] = true
		p2a := fmt.Sprintf("p2a %d %d %d %d %d %s",
			myBallot.id,
			myBallot.n,
			pval.slot,
			pval.req.clientID,
			pval.req.reqNum,
			pval.req.payload)
		rep.send(acc, p2a)
	}
	for rep.isActive {
		payload := <-p2bChan
		acc, ballot := parseP2bPayload(payload)
		if myBallot.eq(ballot) {
			// Accepted :)
			delete(waitfor, acc)
			if len(waitfor) <= int(math.Floor(float64(len(rep.replicas))/2.0)) {
				// pVal is chosen. Broadcast "decision <slot> <clientID> <reqNum> <m>"
				msg := fmt.Sprintf("decision %d %d %d %s",
					pval.slot,
					pval.req.clientID,
					pval.req.reqNum,
					pval.req.payload)
				for learner := range rep.replicas {
					rep.send(learner, msg)
				}
				return
			}
		} else {
			// Pre-empted :(
			preempted <- *ballot
			return
		}
	}
}
