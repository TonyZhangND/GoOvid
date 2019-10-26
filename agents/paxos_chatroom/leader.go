package paxos

// This file describes the states and transitions of a paxos replica that is related
// to its role as an acceptor

import (
	"fmt"
	"math"
	"strings"
	"sync"
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
	p2bMut        *sync.RWMutex          // mutex for p2bOutChans map
}

// Constructor
func (rep *ReplicaAgent) newLeaderState() *leaderState {
	return &leaderState{
		ballotNum:     &ballot{rep.myID, 0},
		active:        false,
		proposals:     make(map[uint64]*proposal),
		proposeInChan: make(chan proposal, bufferSize),
		p2bMut:        new(sync.RWMutex)}
}

// Start running leader thread described in Fig 7 of PMMC
func (rep *ReplicaAgent) runLeader() {
	preemptedInChan := make(chan ballot, bufferSize)          // channel into which scout/cmdr pushes preempted msg
	adoptedInChan := make(chan map[uint64]pValue, bufferSize) // channel into which scout pushes adopted msg
	go rep.spawnScout(
		rep.leader.ballotNum.n,
		preemptedInChan,
		adoptedInChan)
	for rep.isActive {
		rep.debugPrintf("Running Leader loop\n")
		select {
		case prop := <-rep.leader.proposeInChan:
			rep.debugPrintf("Leader received proposal {slot: %d, client: %d, '%s'}\n", prop.slot, prop.req.clientID, prop.req.payload)
			// Handle Propose
			if _, ok := rep.leader.proposals[prop.slot]; !ok {
				// If slot not already used
				rep.leader.proposals[prop.slot] = &prop
				if rep.leader.active {
					cmdP2bOutChan := make(chan string, bufferSize)
					rep.leader.p2bMut.Lock()
					rep.leader.p2bOutChans[prop.slot] = cmdP2bOutChan
					rep.leader.p2bMut.Unlock()
					pval := &pValue{rep.leader.ballotNum.copy(), prop.slot, prop.req}
					go rep.spawnCommander(pval, preemptedInChan, cmdP2bOutChan)
				}
			}
		case pmax := <-adoptedInChan:
			// Handle Adopted
			// pmax is a map of slot->pValue with highest ballot accepted
			rep.debugPrintf("Leader adopted with ballot {%d, %d}\n", rep.leader.ballotNum.id, rep.leader.ballotNum.n)
			for slot, highestAcceptedPVal := range pmax {
				if _, ok := rep.leader.proposals[slot]; ok {
					rep.leader.proposals[slot].req = highestAcceptedPVal.req
				} else {
					prop := &proposal{highestAcceptedPVal.slot, highestAcceptedPVal.req}
					rep.leader.proposals[slot] = prop
				}
			}
			// Spawn commanders for each pval
			for _, prop := range rep.leader.proposals {
				cmdP2bOutChan := make(chan string, bufferSize)
				rep.leader.p2bMut.Lock()
				rep.leader.p2bOutChans[prop.slot] = cmdP2bOutChan
				rep.leader.p2bMut.Unlock()
				pval := &pValue{rep.leader.ballotNum.copy(), prop.slot, prop.req}
				go rep.spawnCommander(pval, preemptedInChan, cmdP2bOutChan)
			}
			rep.leader.active = true
		case bal := <-preemptedInChan:
			// Handle Pre-empted
			rep.debugPrintf("Leader preempted with ballot {%d, %d}\n", rep.leader.ballotNum.id, rep.leader.ballotNum.n)

			// Update my ballot number and spawn scout
			if rep.leader.ballotNum.lt(&bal) {
				rep.leader.active = false
				rep.leader.ballotNum.n = bal.n + 1
			}
			rep.debugPrintf("New ballot {%d, %d}\n", rep.leader.ballotNum.id, rep.leader.ballotNum.n)
			time.Sleep(timeoutDuration * 4)
			preemptedInChan = make(chan ballot, bufferSize)          // channel into which scout/cmdr pushes preempted msg
			adoptedInChan = make(chan map[uint64]pValue, bufferSize) // channel into which scout pushes adopted msg
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

	rep.debugPrintf("Scout spawned for ballot{%d, %d}\n", rep.myID, baln)
	rep.leader.p1bOutChan = make(chan string, bufferSize)
	rep.leader.p2bOutChans = make(map[uint64]chan string) // start a new set of channels
	p1bInChan := rep.leader.p1bOutChan
	waitfor := make(map[c.ProcessID]bool) // set of acceptors from which p1b is pending
	myBallot := &ballot{rep.myID, baln}
	processedPVals := make(map[uint64]pValue)

	for acc := range rep.replicas {
		// Send "p1a <sender> <balNum>"
		waitfor[acc] = true

		go func(acceptor c.ProcessID) {
			p1a := fmt.Sprintf("p1a %d %d", myBallot.id, myBallot.n)
			rep.send(acceptor, p1a)
		}(acc)
	}
	for rep.isActive {
		payload := <-p1bInChan
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
				adoptedOutChan <- processedPVals
				rep.debugPrintf("Scout {%d, %d} killed - adopted\n", rep.myID, baln)
				return
			}
		} else {
			// Pre-empted :(
			preemptedOutChan <- *ballot
			rep.debugPrintf("Scout {%d, %d} killed - preempted\n", rep.myID, baln)
			return
		}
	}
}

// Commander process in Fig 6 of PMMC
func (rep *ReplicaAgent) spawnCommander(
	pval *pValue,
	preemptedOutChan chan ballot,
	p2bInChan chan string) {

	rep.debugPrintf("Commander spawned for pval = {%v, %d, '%s'}\n", *pval.ballot, pval.slot, pval.req.payload)

	waitfor := make(map[c.ProcessID]bool) // set of acceptors from which p2b is pending
	myBallot := pval.ballot

	for acc := range rep.replicas {
		// Send "p2a <balID> <balNum> <slot> <clientID> <reqNum> <m>"
		waitfor[acc] = true
		go func(acceptor c.ProcessID) {
			p2a := fmt.Sprintf("p2a %d %d %d %d %d %s",
				myBallot.id,
				myBallot.n,
				pval.slot,
				pval.req.clientID,
				pval.req.reqNum,
				pval.req.payload)
			rep.send(acceptor, p2a)
			// rep.debugPrintf("Commander sent {%v, %d, '%s'} sent p2a to %d\n", *pval.ballot, pval.slot, pval.req.payload, acc)
		}(acc)
	}
	rep.debugPrintf("Commander {%v, %d, '%s'} sent p2a to all\n", *pval.ballot, pval.slot, pval.req.payload)
	for rep.isActive {
		payload := <-p2bInChan
		acc, _, ballot := parseP2bPayload(payload)
		if myBallot.eq(ballot) {
			// Accepted :)
			delete(waitfor, acc)
			if len(waitfor) <= int(math.Floor(float64(len(rep.replicas))/2.0)) {
				// pVal is chosen. Broadcast "decision <slot> <clientID> <reqNum> <m>"
				rep.debugPrintf("Commander {%v, %d, '%s'} won. Broadcast decision\n", *pval.ballot, pval.slot, pval.req.payload)
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
			rep.debugPrintf("Commander for pval = {%v, %d, %s} preempted. No longer leader\n", pval.ballot, pval.slot, pval.req.payload)
			return
		}
	}
}

// Deliver msg "p1b <accID> <ballotNum.id> <ballotNum.n> <json(accepted pvals)>"
func (rep *ReplicaAgent) handleP1b(request string) {
	rep.leader.p1bOutChan <- strings.SplitN(request, " ", 2)[1]
}

// Deliver msg "p2b <accID> <slot> <ballotNum.id> <ballotNum.n>" Forward it to the right
// commander
func (rep *ReplicaAgent) handleP2b(request string) {
	payload := strings.SplitN(request, " ", 2)[1]
	_, slot, _ := parseP2bPayload(payload)
	rep.leader.p2bMut.RLock()
	c, ok := rep.leader.p2bOutChans[slot]
	rep.leader.p2bMut.RUnlock()
	if ok {
		c <- payload
		rep.debugPrintf("Deliver %s\n", request)
	}
}
