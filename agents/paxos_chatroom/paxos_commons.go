package paxos

import (
	"encoding/json"
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
	bufferSize      = 10000
	commandInterval = 1000 * time.Millisecond
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

func (r *request) eq(other *request) bool {
	return r.hash() == other.hash()
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

// Returns true iff b < other
func (b *ballot) lteq(other *ballot) bool {
	if b.n == other.n {
		return b.id <= other.id
	}
	return b.n < other.n
}

// Returns true iff b = other
func (b *ballot) eq(other *ballot) bool {
	return b.n == other.n && b.id == other.id
}

// Returns a copy of b
func (b *ballot) copy() *ballot {
	return &ballot{b.id, b.n}
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

// Parse "<leaderID> <balNum> <slot> <clientID> <reqNum> <m>" into a pValue
func parsePValue(s string) *pValue {
	sSlice := strings.SplitN(s, " ", 6)
	leaderID, _ := strconv.ParseUint(sSlice[0], 10, 64)
	bNum, _ := strconv.ParseUint(sSlice[1], 10, 64)
	slot, _ := strconv.ParseUint(sSlice[2], 10, 64)
	clientID, _ := strconv.ParseUint(sSlice[3], 10, 64)
	reqNum, _ := strconv.ParseUint(sSlice[4], 10, 64)
	m := sSlice[5]
	return &pValue{
		&ballot{c.ProcessID(leaderID), bNum},
		slot,
		&request{c.ProcessID(clientID), reqNum, m}}
}

// Parse "<accID> <ballotNum.id> <ballotNum.n> <json.Marshal(accepted)>"
// into (accID, ballot, map of slot->pValue)
func parseP1bPayload(s string) (c.ProcessID, *ballot, map[uint64]*pValue) {
	sSlice := strings.SplitN(s, " ", 4)
	accID, _ := strconv.ParseUint(sSlice[0], 10, 64)
	bID, _ := strconv.ParseUint(sSlice[1], 10, 64)
	bn, _ := strconv.ParseUint(sSlice[2], 10, 64)
	pVals := make(map[uint64]*pValue)

	// Parse and populate pVals
	var dat map[string]interface{}
	if err := json.Unmarshal([]byte(sSlice[3]), &dat); err != nil {
		fmt.Println("HELP")
		panic(err)
	}
	for k, v := range dat {
		slot, _ := strconv.ParseUint(k, 10, 64)
		pVal := parsePValue(v.(string))
		pVals[slot] = pVal
	}
	return c.ProcessID(accID), &ballot{c.ProcessID(bID), bn}, pVals
}

// Parse "<accID> <slot> <ballotNum.id> <ballotNum.n>"
// into (accID, ballot)
func parseP2bPayload(s string) (c.ProcessID, uint64, *ballot) {
	sSlice := strings.SplitN(s, " ", 4)
	accID, _ := strconv.ParseUint(sSlice[0], 10, 64)
	slot, _ := strconv.ParseUint(sSlice[1], 10, 64)
	bID, _ := strconv.ParseUint(sSlice[2], 10, 64)
	bn, _ := strconv.ParseUint(sSlice[3], 10, 64)
	return c.ProcessID(accID), slot, &ballot{c.ProcessID(bID), bn}
}
