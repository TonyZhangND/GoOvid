package main

import (
	"fmt"
	"os"
	"sync"
)

type (
	processID    uint16
	stateTracker struct {
		status map[processID]bool
		sync.RWMutex
	}
)

const (
	basePort = 3000
)

// Constructor for state tracker
func newStateTracker() stateTracker {
	return stateTracker{status: make(map[processID]bool)}
}

// Add process pid to the state tracker
func (st *stateTracker) addProcess(pid processID) {
	st.RLock()
	_, ok := st.status[pid]
	st.Unlock()
	if ok {
		fmt.Printf("Error: process %v already exist in failure detector", pid)
		os.Exit(1)
	}
	st.Lock()
	st.status[pid] = false
	st.Unlock()
}

// Mark a process as down in st
func (st *stateTracker) markAsDown(pid processID) {
	st.RLock()
	_, ok := st.status[pid]
	st.Unlock()
	if !ok {
		fmt.Printf("Error: process %v does not exist in failure detector", pid)
		os.Exit(1)
	}
	st.Lock()
	st.status[pid] = false
	st.Unlock()
}

// Mark a process as up in st
func (st *stateTracker) markAsUp(pid processID) {
	st.RLock()
	_, ok := st.status[pid]
	st.Unlock()
	if !ok {
		fmt.Printf("Error: process %v does not exist in failure detector", pid)
		os.Exit(1)
	}
	st.Lock()
	st.status[pid] = true
	st.Unlock()
}

// Returns a slice containing the list of live processes in st
func (st *stateTracker) getAlive() []processID {
	result := make([]processID, 0)
	st.RLock()
	for pid, up := range st.status { // find the nodes that are up
		if up {
			result = append(result, pid)
		}
	}
	defer st.RUnlock()
	return result
}
