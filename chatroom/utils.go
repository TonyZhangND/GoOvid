package main

import (
	"fmt"
	"os"
	"sync"
)

type processID uint16

// A connTracker ct is a thread-safe map from processID to a connHandler object.
// It is initialized as ct[p]=nil for all known processes p.
// For each process p, when a connHandler c is established with it,
// we mark ct[p] = c.
type connTracker struct {
	tracker map[processID]*connHandler
	sync.RWMutex
}

const (
	basePort = 3000
)

// Constructor for connection tracker
// It takes a slice of all known process IDs, and initializes a
// connTracker ct with ct[p]=nil for all p in knownProcesses.
func newConnTracker(knownProcesses []processID) connTracker {
	t := make(map[processID]*connHandler)
	for _, pid := range knownProcesses {
		t[pid] = nil
	}
	return connTracker{tracker: t}
}

// // Adds process pid to the state tracker
// func (st *connTracker) trackProcess(pid processID) {
// 	st.RLock()
// 	_, ok := st.tracker[pid]
// 	st.RUnlock()
// 	if ok {
// 		fmt.Printf("Error: process %v already exist in the connection tracker", pid)
// 		os.Exit(1)
// 	}
// 	st.Lock()
// 	st.tracker[pid] = nil
// 	st.Unlock()
// }

// Marks a process as down in ct and de-registers its connHandler object
func (ct *connTracker) markAsDown(pid processID) {
	ct.RLock()
	handler, ok := ct.tracker[pid]
	if handler == nil {
		defer ct.RUnlock()
		return
	}
	ct.RUnlock()
	if !ok {
		fmt.Printf("Error: process %v does not exist in failure detector", pid)
		os.Exit(1)
	}
	ct.Lock()
	ct.tracker[pid] = nil
	ct.Unlock()
}

// Marks a process as up in ct and register its connHandler object
func (ct *connTracker) markAsUp(pid processID, handler *connHandler) {
	ct.RLock()
	conn, ok := ct.tracker[pid]
	if conn != nil {
		defer ct.RUnlock()
		return
	}
	ct.RUnlock()
	if !ok {
		fmt.Printf("Error: process %v does not exist in failure detector", pid)
		os.Exit(1)
	}
	ct.Lock()
	ct.tracker[pid] = handler
	ct.Unlock()
}

// Returns true iff process pid is up in ct
func (ct *connTracker) isUp(pid processID) bool {
	ct.RLock()
	handler, ok := ct.tracker[pid]
	if !ok {
		fmt.Printf("Error: process %v does not exist in failure detector", pid)
		os.Exit(1)
	}
	defer ct.RUnlock()
	return handler != nil
}

// Returns a slice containing the list of up processes in ct
func (ct *connTracker) getAlive() []processID {
	result := make([]processID, 0)
	ct.RLock()
	for pid, handler := range ct.tracker {
		// find the nodes that are up
		if handler != nil {
			result = append(result, pid)
		}
	}
	defer ct.RUnlock()
	return result
}

// Returns a slice containing the list of down processes in ct
func (ct *connTracker) getDead() []processID {
	result := make([]processID, 0)
	ct.RLock()
	for pid, handler := range ct.tracker {
		if handler == nil {
			result = append(result, pid)
		}
	}
	defer ct.RUnlock()
	return result
}

// Sends msg on all channels.
// Applies Ovid message format and headers
func (ct *connTracker) broadcast(msg string) {
	s := fmt.Sprintf("msg %v %s\n", myPhysID, msg)
	ct.RLock()
	for _, handler := range ct.tracker {
		if handler != nil {
			handler.send(s)
		}
	}
	ct.RUnlock()
	return
}
