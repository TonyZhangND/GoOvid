package paxos

import (
	"encoding/json"
	"fmt"
	"strings"
)

// This file describes the states and transitions of a paxos replica that is related
// to its role as an acceptor

type acceptorState struct {
	ballotNum *ballot
	accepted  map[uint64]string
	// accepted is map of slot to p2aPayload (i.e. string describing pValue)
	// "<leaderID> <bNum> <slot> <clientID> <reqNum> <m>"
}

// Constructor
func (rep *ReplicaAgent) newAcceptorState() *acceptorState {
	return &acceptorState{accepted: make(map[uint64]string)}
}

// Handle msg "p1a <sender> <balNum>"
func (rep *ReplicaAgent) handleP1a(s string) {
	payload := strings.SplitN(s, " ", 2)[1]
	leaderID, bNum := parseP1aPayload(payload)
	newBallot := &ballot{leaderID, bNum}
	if rep.acceptor.ballotNum == nil || rep.acceptor.ballotNum.lt(newBallot) {
		rep.acceptor.ballotNum = newBallot
	}
	// Respond with "p1b <myID> <ballotNum.id> <ballotNum.n> <json.Marshal(accepted)>"
	m, _ := json.Marshal(rep.acceptor.accepted)
	response := fmt.Sprintf("p1b %d %d %d %s",
		rep.myID,
		rep.acceptor.ballotNum.id,
		rep.acceptor.ballotNum.n,
		m)
	rep.send(leaderID, response)
	rep.debugPrintf("Acceptor sent p1b to %d\n", leaderID)
}

// Handle msg "p2a <balID> <balNum> <slot> <clientID> <reqNum> <m>"
func (rep *ReplicaAgent) handleP2a(s string) {
	sSlice := strings.SplitN(s, " ", 2)
	pval := parsePValue(sSlice[1])
	if rep.acceptor.ballotNum.id == pval.ballot.id &&
		rep.acceptor.ballotNum.n == pval.ballot.n {
		// Accept pVal if I did not promise some higher ballot
		pValStr := sSlice[1]
		rep.acceptor.accepted[pval.slot] = pValStr
	}
	// Respond with "p2b <myID> <slot> <ballotNum.id> <ballotNum.n>"
	response := fmt.Sprintf("p2b %d %d %d %d",
		rep.myID,
		pval.slot,
		rep.acceptor.ballotNum.id,
		rep.acceptor.ballotNum.n)
	rep.send(pval.ballot.id, response)
}
