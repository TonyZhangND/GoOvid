package paxos

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	c "github.com/TonyZhangND/GoOvid/commons"
)

// This file describes the states and transitions of a paxos replica that is related
// to its role as an acceptor

type acceptorState struct {
	ballotNum *ballot
	accepted  map[uint64]string
	// accepted is map of slot to
	// "<repID> <bNum> <slot> <clientID> <reqNum> <m>"
}

// Constructor
func newAcceptorState() *acceptorState {
	return &acceptorState{}
}

// Handle msg "p1a <sender> <balNum>"
func (rep *ReplicaAgent) handleP1a(s string) {
	sSlice := strings.SplitN(s, " ", 3)
	leaderID, _ := strconv.ParseUint(sSlice[1], 10, 64)
	bNum, _ := strconv.ParseUint(sSlice[2], 10, 64)
	newBallot := &ballot{c.ProcessID(leaderID), bNum}
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
	rep.send(c.ProcessID(leaderID), response)
}

// p2a "<sender> <balNum> <slot> <clientID> <reqNum> <m>"
func (rep *ReplicaAgent) handleP2a(s string) {

}
