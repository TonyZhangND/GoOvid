package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

// A linkManager lm is a thread-safe map from processID to a link object.
// It is initialized as lm[p]=nil for all known processes p.
// For each process p, when a link l is established with it,
// we mark lm[p] = l.
type linkManager struct {
	manager       map[processID]*link
	masterConn    net.Conn
	serverOutChan chan string // used to stream messages to main server loop
	masterOutChan chan string // used to stream messages to main server loop
	sync.RWMutex
}

// Constructor for linkManager
// It takes a slice of all known process IDs, and initializes a
// connTracker lm with lm[p]=nil for all p in knownProcesses.
func newLinkManager(knownProcesses []processID, sOutChan chan string,
	mOutChan chan string) *linkManager {
	t := make(map[processID]*link)
	for _, pid := range knownProcesses {
		t[pid] = nil
	}
	return &linkManager{manager: t, serverOutChan: sOutChan, masterOutChan: mOutChan}
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
		errMsg := fmt.Sprintf(
			"Process %v does not exist in failure detector", pid)
		fatalError(errMsg)
	}
	lm.Lock()
	lm.manager[pid] = nil
	lm.Unlock()
}

// Marks a process as up in lm and register its link object
func (lm *linkManager) markAsUp(pid processID, handler *link) {
	lm.RLock()
	link, ok := lm.manager[pid]
	if link != nil && pid != myPhysID {
		errMsg := fmt.Sprintf("Link to %v already established!", pid)
		fatalError(errMsg)
	}
	lm.RUnlock()
	if !ok {
		errMsg := fmt.Sprintf(
			"Process %v does not exist in failure detector", pid)
		fatalError(errMsg)
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
		errMsg := fmt.Sprintf(
			"Process %v does not exist in failure detector", pid)
		fatalError(errMsg)
	}
	defer lm.RUnlock()
	return link != nil
}

// Returns a slice containing the list of up processes in lm
func (lm *linkManager) getAlive() []processID {
	// The comparator function for sorting a slice of processIDs
	result := make([]processID, 0)
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

// sendToMaster sends msg string to the master
func (lm *linkManager) sendToMaster(msg string) {
	_, err := lm.masterConn.Write([]byte(msg + "\n"))
	if err != nil {
		errMsg := fmt.Sprintf("Can't send msg '%v' to master: %v",
			msg, err)
		fatalError(errMsg)
	}
}

// Dials for new connections to all pid <= my pid
func (lm *linkManager) dialForConnections() {
	debugPrintln("Dialing for peer connections")
	for shouldRun {
		down := linkMgr.getDead()
		for _, pid := range down {
			if pid <= myPhysID && !linkMgr.isUp(pid) {
				dialingAddr := fmt.Sprintf("%s:%d", gridIP, basePort+pid)
				c, err := net.DialTimeout("tcp", dialingAddr,
					20*time.Millisecond)
				if err == nil {
					l := newLinkKnownOther(c, pid)
					go l.handleConnection()
				}
			}
		}
		time.Sleep(1 * time.Second)
	}
}

// Listens and establishes new connections
func (lm *linkManager) listenForConnections() {
	debugPrintln("Listening for peer connections")
	listenerAddr := fmt.Sprintf("%s:%d", gridIP, basePort+myPhysID)
	l, err := net.Listen("tcp", listenerAddr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer l.Close()
	for shouldRun {
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		l := newLink(c)
		go l.handleConnection()
	}
}

// Connect to and handle the master
func (lm *linkManager) connectAndHandleMaster() {
	// listen for master on the master address
	masterAddr := fmt.Sprintf("%s:%d", masterIP, masterPort)
	debugPrintln("Listening for master connecting on " + masterAddr)
	mstrListener, _ := net.Listen("tcp", masterAddr)
	mstrConn, _ := mstrListener.Accept()
	lm.masterConn = mstrConn
	debugPrintln("Accepted master connection")
	// main loop: process commands from master as async goroutine
	// This allows main server loop to progress only after the master has connected
	go func() {
		defer lm.masterConn.Close()
		connReader := bufio.NewReader(lm.masterConn)
		for shouldRun {
			data, err := connReader.ReadString('\n')
			if err != nil {
				fatalError("Broken connection from master")
				break
			}
			lm.masterOutChan <- data
		}
	}()
}

func (lm *linkManager) run() {
	go lm.dialForConnections()
	go lm.listenForConnections()
	lm.connectAndHandleMaster()
}
