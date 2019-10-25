package paxos

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	c "github.com/TonyZhangND/GoOvid/commons"
)

const (
	pingInterval    = 100 * time.Microsecond
	sleepDuration   = 100 * time.Millisecond
	timeoutDuration = 1000 * time.Millisecond
)

var wg sync.WaitGroup

// a request describes a client request
type request struct {
	clientID c.ProcessID
	reqNum   uint64
	payload  string
}

func (r *request) hash() string {
	return fmt.Sprintf("%v", *r)
}

// a proposal describes a (slot, request) pair
type proposal struct {
	slot uint64 // slot number
	req  *request
}

func (p *proposal) hash() string {
	return fmt.Sprintf("%d : %s", p.slot, p.req.hash())
}

type ballot struct {
	id c.ProcessID
	n  uint64
}

// Returns true iff b < other
func (b *ballot) lt(other *ballot) bool {
	if b.n == other.n {
		return b.id < other.id
	}
	return b.n < other.n
}

type pValue struct {
	ballot *ballot
	slot   uint64
	req    *request
}

// Parse "<sender> <balNum>" and return sender, balNum
func parseP1aPayload(s string) (c.ProcessID, uint64) {
	sSlice := strings.SplitN(s, " ", 2)
	leaderID, _ := strconv.ParseUint(sSlice[0], 10, 64)
	bNum, _ := strconv.ParseUint(sSlice[1], 10, 64)
	return c.ProcessID(leaderID), bNum
}

// Parse "<leaderID> <balNum> <slot> <clientID> <reqNum> <m>"
func parseP2aPayload(s string) (c.ProcessID, uint64, uint64, c.ProcessID, uint64, string) {
	sSlice := strings.SplitN(s, " ", 6)
	leaderID, _ := strconv.ParseUint(sSlice[0], 10, 64)
	bNum, _ := strconv.ParseUint(sSlice[1], 10, 64)
	slot, _ := strconv.ParseUint(sSlice[2], 10, 64)
	clientID, _ := strconv.ParseUint(sSlice[3], 10, 64)
	reqNum, _ := strconv.ParseUint(sSlice[4], 10, 64)
	m := sSlice[5]
	return c.ProcessID(leaderID), bNum, slot, c.ProcessID(clientID), reqNum, m
}
