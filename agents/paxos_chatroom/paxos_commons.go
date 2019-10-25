package paxos

import (
	"fmt"
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

// // a decision describes a (slot, request) pair
// type decision struct {
// 	slot uint64 // slot number
// 	req  *request
// }
