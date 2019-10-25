package paxos

// This file describes the states and transitions of a paxos replica that is related
// to its role as an acceptor

import (
	"fmt"
	"math"
	"strings"
	"time"

	c "github.com/TonyZhangND/GoOvid/commons"
)

type leaderState struct {
	ballotNum     *ballot
	active        bool
	proposals     map[uint64]*proposal   // given k->*v, k is a hash of v
	proposeInChan chan proposal          // channel into which replica pushes proposals
	p1bOutChan    chan string            // channel into which leader pushes p1b to scout
	p2bOutChans   map[uint64]chan string // channels into which leader pushes p1b to commanders
}

// Constructor
func (rep *ReplicaAgent) newLeaderState() *leaderState {
	return &leaderState{
		ballotNum:     &ballot{rep.myID, 0},
		active:        false,
		proposals:     make(map[uint64]*proposal),
		proposeInChan: make(chan proposal, len(rep.replicas))}
}

// Start running leader thread described in Fig 7 of PMMC
func (rep *ReplicaAgent) runLeader() {
	preemptedInChan := make(chan ballot)          // channel into which scout/cmdr pushes preempted msg
	adoptedInChan := make(chan map[uint64]pValue) // channel into which scout pushes adopted msg
	go rep.spawnScout(
		rep.leader.ballotNum.n,
		preemptedInChan,
		adoptedInChan)
	for rep.isActive {
		rep.debugPrintf("Running Leader loop\n")
		select {
		case prop := <-rep.leader.proposeInChan:
			// Handle Propose
			if _, ok := rep.leader.proposals[prop.slot]; !ok {
				// If slot not already used
				rep.leader.proposals[prop.slot] = &prop
				if rep.leader.active {
					cmdP2bOutChan := make(chan string)
					rep.leader.p2bOutChans[prop.slot] = cmdP2bOutChan
					pval := &pValue{rep.leader.ballotNum.copy(), prop.slot, prop.req}
					go rep.spawnCommander(pval, preemptedInChan, cmdP2bOutChan)
				}
			}
		case pmax := <-adoptedInChan:
			// Handle Adopted
			// pmax is a map of slot->pValue with highest ballot accepted
			rep.debugPrintf("I am leader\n")
			for slot, highestAcceptedPVal := range pmax {
				rep.leader.proposals[slot].req = highestAcceptedPVal.req
			}
			// Spawn commanders for each pval
			for _, prop := range rep.leader.proposals {
				cmdP2bOutChan := make(chan string, bufferSize)
				rep.leader.p2bOutChans[prop.slot] = cmdP2bOutChan
				pval := &pValue{rep.leader.ballotNum.copy(), prop.slot, prop.req}
				go rep.spawnCommander(pval, preemptedInChan, cmdP2bOutChan)
			}
			rep.leader.active = true
		case bal := <-preemptedInChan:
			// Handle Pre-empted
			// Update my ballot number and spawn scout
			if rep.leader.ballotNum.lt(&bal) {
				rep.leader.active = false
				rep.leader.ballotNum.n = bal.n + 1
			}
			time.Sleep(timeoutDuration * 4)
			preemptedInChan = make(chan ballot)          // channel into which scout/cmdr pushes preempted msg
			adoptedInChan = make(chan map[uint64]pValue) // channel into which scout pushes adopted msg
			go rep.spawnScout(
				rep.leader.ballotNum.n,
				preemptedInChan,
				adoptedInChan)
		}
	}
}

// Scout process in Fig 6 of PMMC
func (rep *ReplicaAgent) spawnScout(
	baln uint64,
	preemptedOutChan chan ballot, // channel into which scout pushes preempted msg
	adoptedOutChan chan map[uint64]pValue) { // channel into which scout pushes adopted msg

	rep.leader.p1bOutChan = make(chan string, bufferSize)
	p1bInChan := rep.leader.p1bOutChan
	rep.debugPrintf("Read channel %v\n", p1bInChan)
	rep.debugPrintf("Scout spawned\n")
	waitfor := make(map[c.ProcessID]bool) // set of acceptors from which p1b is pending
	myBallot := &ballot{rep.myID, baln}
	processedPVals := make(map[uint64]pValue)
	for acc := range rep.replicas {
		// Send "p1a <sender> <balNum>"
		waitfor[acc] = true
		p1a := fmt.Sprintf("p1a %d %d", myBallot.id, myBallot.n)
		rep.send(acc, p1a)
	}
	rep.debugPrintf("Scout entering loop\n")
	for rep.isActive {
		payload := <-p1bInChan
		rep.debugPrintf("JE::LP\n")
		acc, ballot, pVals := parseP1bPayload(payload)
		rep.debugPrintf("Scout received p1b from %d\n", acc)
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
				adoptedOutChan <- processedPVals
				rep.debugPrintf("Scout killed - adopted\n")
				return
			}
		} else {
			// Pre-empted :(
			preemptedOutChan <- *ballot
			rep.debugPrintf("Scout killed - preempted\n")
			return
		}
	}
}

// Commander process in Fig 6 of PMMC
func (rep *ReplicaAgent) spawnCommander(
	pval *pValue,
	preemptedOutChan chan ballot,
	p2bInChan chan string) {

	rep.leader.p2bOutChans = make(map[uint64]chan string) // start a new set of channels
	waitfor := make(map[c.ProcessID]bool)                 // set of acceptors from which p2b is pending
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
		payload := <-p2bInChan
		acc, _, ballot := parseP2bPayload(payload)
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
			preemptedOutChan <- *ballot
			return
		}
	}
}

// Deliver msg "p1b <accID> <ballotNum.id> <ballotNum.n> <json(accepted pvals)>"
func (rep *ReplicaAgent) handleP1b(request string) {
	rep.debugPrintf("Write channel %v\n", rep.leader.p1bOutChan)
	rep.leader.p1bOutChan <- strings.SplitN(request, " ", 2)[1]
	rep.debugPrintf("Received p1b from %s\n", request)
}

// Deliver msg "p2b <accID> <slot> <ballotNum.id> <ballotNum.n>" Forward it to the right
// commander
func (rep *ReplicaAgent) handleP2b(request string) {
	payload := strings.SplitN(request, " ", 2)[1]
	_, slot, _ := parseP2bPayload(payload)
	if c, ok := rep.leader.p2bOutChans[slot]; ok {
		c <- payload
	}
}
