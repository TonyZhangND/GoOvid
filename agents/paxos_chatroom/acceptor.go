package paxos

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

// This file describes the states and transitions of a paxos replica that is related
// to its role as an acceptor

type acceptorState struct {
	ballotNum *ballot
	accepted  map[uint64]string
	amut      *sync.RWMutex // Mutex for accepted map
	// accepted is map of slot to p2aPayload (i.e. string describing pValue)
	// "<leaderID> <bNum> <slot> <clientID> <reqNum> <m>"
}

// Constructor
func (rep *ReplicaAgent) newAcceptorState() *acceptorState {
	return &acceptorState{
		accepted: make(map[uint64]string),
		amut:     new(sync.RWMutex)}
}

// Handle msg "p1a <sender> <balNum>"
func (rep *ReplicaAgent) handleP1a(s string) {
	rep.debugPrintf("Receive %s\n", s)
	payload := strings.SplitN(s, " ", 2)[1]
	leaderID, bNum := parseP1aPayload(payload)
	newBallot := &ballot{leaderID, bNum}
	if rep.acceptor.ballotNum == nil || rep.acceptor.ballotNum.lt(newBallot) {
		rep.acceptor.ballotNum = newBallot
	}
	// Respond with "p1b <myID> <ballotNum.id> <ballotNum.n> <json.Marshal(accepted)>"
	rep.acceptor.amut.RLock()
	m, _ := json.Marshal(rep.acceptor.accepted)
	rep.acceptor.amut.RUnlock()
	response := fmt.Sprintf("p1b %d %d %d %s",
		rep.myID,
		rep.acceptor.ballotNum.id,
		rep.acceptor.ballotNum.n,
		m)
	rep.send(leaderID, response)
	rep.debugPrintf("Sent %s to %d\n", response, leaderID)
}

// Handle msg "p2a <balID> <balNum> <slot> <clientID> <reqNum> <m>"
func (rep *ReplicaAgent) handleP2a(s string) {
	rep.debugPrintf("Receive p2a %s\n", s)
	sSlice := strings.SplitN(s, " ", 2)
	pval := parsePValue(sSlice[1])
	if rep.acceptor.ballotNum == nil ||
		(rep.acceptor.ballotNum.id <= pval.ballot.id &&
			rep.acceptor.ballotNum.n <= pval.ballot.n) {
		// Accept pVal if I did not promise some higher ballot
		pValStr := sSlice[1]
		newBal := &ballot{pval.ballot.id, pval.ballot.n}
		rep.acceptor.ballotNum = newBal
		rep.acceptor.amut.Lock()
		rep.acceptor.accepted[pval.slot] = pValStr
		rep.acceptor.amut.Unlock()
	}
	// Respond with "p2b <myID> <slot> <ballotNum.id> <ballotNum.n>"
	response := fmt.Sprintf("p2b %d %d %d %d",
		rep.myID,
		pval.slot,
		rep.acceptor.ballotNum.id,
		rep.acceptor.ballotNum.n)
	rep.send(pval.ballot.id, response)
}
