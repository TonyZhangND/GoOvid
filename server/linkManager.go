package server

// This file contains the definition and methods of the linkManager object.
// A linkManager acts as a networking interface for an Ovid server.
// It manages all active links, and contains methods to query the state
// of the network.

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

// A linkManager lm implements a thread-safe map from processID
// to a link object. It is initialized as lm[p]=nil for all known
// processes p. For each process p != myPhysId, when a link l is
// established with it, we mark lm[p] = l.
// It also maintains the connection with the master program
// Note: always lm[myPhysId] = nil, since a server does not need a link
// with itself.
type linkManager struct {
	manager       map[processID]*link
	masterConn    net.Conn    // connection with the master program
	serverOutChan chan string // used to stream server messages to main server loop
	masterOutChan chan string // used to stream master messages to main server loop
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
		FatalError(errMsg)
	}
	lm.Lock()
	lm.manager[pid] = nil
	lm.Unlock()
}

// Marks a process as up in lm and registers its link object
func (lm *linkManager) markAsUp(pid processID, handler *link) {
	lm.RLock()
	link, ok := lm.manager[pid]
	if link != nil && pid != myPhysID {
		errMsg := fmt.Sprintf("Link to %v already established!", pid)
		FatalError(errMsg)
	}
	lm.RUnlock()
	if !ok {
		errMsg := fmt.Sprintf(
			"Process %v does not exist in failure detector", pid)
		FatalError(errMsg)
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
		FatalError(errMsg)
	}
	defer lm.RUnlock()
	return link != nil
}

// Returns a slice containing the list of up processes in lm
func (lm *linkManager) getAllUp() []processID {
	// The comparator function for sorting a slice of processIDs
	result := make([]processID, 0)
	lm.RLock()
	for pid, link := range lm.manager {
		// find the nodes that are up
		if link != nil {
			result = append(result, pid)
		}
	}
	result = append(result, myPhysID) // Cogito ergo sum
	defer lm.RUnlock()
	return result
}

// Returns a slice containing the list of down processes in lm
func (lm *linkManager) getAllDown() []processID {
	result := make([]processID, 0)
	lm.RLock()
	for pid, link := range lm.manager {
		if link == nil && pid != myPhysID { // cogito ergo sum
			result = append(result, pid)
		}
	}
	defer lm.RUnlock()
	return result
}

// Sends msg on all channels.
// Applies Ovid message format and headers
func (lm *linkManager) broadcast(msg string) {
	lm.serverOutChan <- msg // first send to myself
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

// Sends msg string to the master
func (lm *linkManager) sendToMaster(msg string) {
	_, err := lm.masterConn.Write([]byte(msg + "\n"))
	if err != nil {
		errMsg := fmt.Sprintf("Can't send msg '%v' to master: %v",
			msg, err)
		FatalError(errMsg)
	}
}

// Dials for new connections to all pid < my pid
func (lm *linkManager) dialForConnections() {
	DebugPrintln("Dialing for peer connections")
	for shouldRun {
		down := linkMgr.getAllDown()
		for _, pid := range down {
			if pid < myPhysID && !linkMgr.isUp(pid) {
				dialingAddr := fmt.Sprintf("%s:%d", gridIP, basePort+pid)
				c, err := net.DialTimeout("tcp", dialingAddr,
					20*time.Millisecond)
				if err == nil {
					l := newLinkKnownOther(c, pid, lm.serverOutChan)
					go l.handleConnection()
				}
			}
		}
		time.Sleep(1 * time.Second)
	}
}

// Listens and establishes new connections
func (lm *linkManager) listenForConnections() {
	DebugPrintln("Listening for peer connections")
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
		l := newLink(c, lm.serverOutChan)
		go l.handleConnection()
	}
}

// Connects to and handles the master
func (lm *linkManager) connectAndHandleMaster() {
	// listen for master on the master address
	masterAddr := fmt.Sprintf("%s:%d", masterIP, masterPort)
	DebugPrintln("Listening for master connecting on " + masterAddr)
	mstrListener, err := net.Listen("tcp", masterAddr)
	if err != nil {
		FatalError("Error connecting to master")
	}
	mstrConn, err := mstrListener.Accept()
	if err != nil {
		FatalError("Error connecting to master")
	}
	lm.masterConn = mstrConn
	DebugPrintln("Accepted master connection")
	// main loop: process commands from master as async goroutine
	go func() {
		defer lm.masterConn.Close()
		connReader := bufio.NewReader(lm.masterConn)
		for shouldRun {
			data, err := connReader.ReadString('\n')
			if err != nil {
				FatalError("Broken connection from master")
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
	// Rather than running connectAndHandleMaster() as an async goroutine,
	// we defer the goroutine to within connectAndHandleMaster(). This forces
	// the main server loop to progress only after the master has connected.
}
