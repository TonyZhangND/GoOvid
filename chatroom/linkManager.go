package main

import (
	"fmt"
	"os"
	"sort"
	"sync"
)

// A linkManager lm is a thread-safe map from processID to a link object.
// It is initialized as lm[p]=nil for all known processes p.
// For each process p, when a link l is established with it,
// we mark lm[p] = l.
type linkManager struct {
	manager map[processID]*link
	sync.RWMutex
}

const (
	basePort = 3000
)

// Constructor for linkManager
// It takes a slice of all known process IDs, and initializes a
// connTracker lm with lm[p]=nil for all p in knownProcesses.
func newLinkManager(knownProcesses []processID) *linkManager {
	t := make(map[processID]*link)
	for _, pid := range knownProcesses {
		t[pid] = nil
	}
	return &linkManager{manager: t}
}

// Marks a process as down in lm and de-registers its link object
func (lm *linkManager) markAsDown(pid processID) {
	lm.RLock()
	link, ok := lm.manager[pid]
	if link == nil {
		defer lm.RUnlock()
		return
	}
	lm.RUnlock()
	if !ok {
		fmt.Printf("Error: process %v does not exist in failure detector", pid)
		os.Exit(1)
	}
	lm.Lock()
	lm.manager[pid] = nil
	lm.Unlock()
}

// Marks a process as up in lm and register its link object
func (lm *linkManager) markAsUp(pid processID, handler *link) {
	lm.RLock()
	link, ok := lm.manager[pid]
	if link != nil {
		defer lm.RUnlock()
		return
	}
	lm.RUnlock()
	if !ok {
		fmt.Printf("Error: process %v does not exist in failure detector", pid)
		os.Exit(1)
	}
	lm.Lock()
	lm.manager[pid] = handler
	lm.Unlock()
}

// Returns true iff process pid is up in lm
func (lm *linkManager) isUp(pid processID) bool {
	lm.RLock()
	link, ok := lm.manager[pid]
	if !ok {
		fmt.Printf("Error: process %v does not exist in failure detector", pid)
		os.Exit(1)
	}
	defer lm.RUnlock()
	return link != nil
}

// Returns a slice containing the list of up processes in lm
func (lm *linkManager) getAlive() []processID {
	// The comparator function for sorting a slice of processIDs
	result := make([]processID, 0)
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	lm.RLock()
	for pid, link := range lm.manager {
		// find the nodes that are up
		if link != nil {
			result = append(result, pid)
		}
	}
	defer lm.RUnlock()
	return result
}

// Returns a slice containing the list of down processes in lm
func (lm *linkManager) getDead() []processID {
	result := make([]processID, 0)
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	lm.RLock()
	for pid, link := range lm.manager {
		if link == nil {
			result = append(result, pid)
		}
	}
	defer lm.RUnlock()
	return result
}

// Sends msg on all channels.
// Applies Ovid message format and headers
func (lm *linkManager) broadcast(msg string) {
	s := fmt.Sprintf("msg %v %s\n", myPhysID, msg)
	lm.RLock()
	for _, link := range lm.manager {
		if link != nil {
			link.send(s)
		}
	}
	lm.RUnlock()
	return
}
