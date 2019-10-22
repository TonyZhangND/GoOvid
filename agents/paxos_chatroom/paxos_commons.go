package paxos

import (
	"sync"
	"time"
)

const (
	sleepDuration   = 100 * time.Millisecond
	timeoutDuration = 5000 * time.Millisecond
)

var wg sync.WaitGroup
