package main

import (
	"fmt"
	"net"
	"os"
	"sync"
)

type processID uint16

// A connTracker ct is a map from processID to a net.Conn object.
// It is initialized as ct[p]=nil for all known processes p.
// For each process p, when a net.Conn object c is established with it,
// we mark ct[p] = c.
type connTracker struct {
	tracker map[processID]net.Conn
	sync.RWMutex
}

const (
	basePort = 3000
)

// Constructor for connection tracker
// It takes a slice of all known process IDs, and initializes a
// connTracker ct with ct[p]=nil for all p in knownProcesses.
func newConnTracker(knownProcesses []processID) connTracker {
	t := make(map[processID]net.Conn)
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

// Marks a process as down in ct
func (ct *connTracker) markAsDown(pid processID) {
	ct.RLock()
	conn, ok := ct.tracker[pid]
	if conn == nil {
		defer ct.RUnlock()
		return
	}
	ct.RUnlock()
	if !ok {
		fmt.Printf("Error: process %v does not exist in failure detector", pid)
		os.Exit(1)
	}
	conn.Close()
	ct.Lock()
	ct.tracker[pid] = nil
	ct.Unlock()
}

// Marks a process as up in ct and register its net.Conn object
func (ct *connTracker) markAsUp(pid processID, connection net.Conn) {
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
	ct.tracker[pid] = connection
	ct.Unlock()
}

// Returns true iff process pid is up in ct
func (ct *connTracker) isUp(pid processID) bool {
	ct.RLock()
	conn, ok := ct.tracker[pid]
	if !ok {
		fmt.Printf("Error: process %v does not exist in failure detector", pid)
		os.Exit(1)
	}
	defer ct.RUnlock()
	return conn != nil
}

// Returns a slice containing the list of up processes in st
func (ct *connTracker) getAlive() []processID {
	result := make([]processID, 0)
	ct.RLock()
	for pid, conn := range ct.tracker {
		// find the nodes that are up
		if conn != nil {
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
	for pid, conn := range ct.tracker {
		if conn == nil {
			result = append(result, pid)
		}
	}
	defer ct.RUnlock()
	return result
}

// Returns a true iff pid is a known process in ct
func (ct *connTracker) isKnown(pid processID) bool {
	ct.RLock()
	defer ct.RUnlock()
	_, ok := ct.tracker[pid]
	if ok {
		return true
	}
	return false
}
