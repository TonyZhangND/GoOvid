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

	c "github.com/TonyZhangND/GoOvid/commons"
)

// A linkManager lm implements a thread-safe map from processID
// to a link object. It is initialized as lm[p]=nil for all known
// processes p. For each process p != myPhysId, when a link l is
// established with it, we mark lm[p] = l.
// It also maintains the connection with the master program
// Note: always lm[myPhysId] = nil, since a server does not need a link
// with itself.
type linkManager struct {
	manager       map[c.ProcessID]*link
	masterConn    net.Conn    // connection with the master program
	serverOutChan chan string // used to stream server messages to main server loop
	masterOutChan chan string // used to stream master messages to main server loop
	sync.RWMutex
}

// Constructor for linkManager
// It takes a slice of all known process IDs, and initializes a
// connTracker lm with lm[p]=nil for all p in knownProcesses.
func newLinkManager(knownProcesses []c.ProcessID, sOutChan chan string,
	mOutChan chan string) *linkManager {
	t := make(map[c.ProcessID]*link)
	for _, pid := range knownProcesses {
		t[pid] = nil
	}
	return &linkManager{manager: t, serverOutChan: sOutChan, masterOutChan: mOutChan}
}

// Marks a process as down in lm and de-registers its link object
func (lm *linkManager) markAsDown(pid c.ProcessID) {
	lm.RLock()
	link, ok := lm.manager[pid]
	if link == nil {
		defer lm.RUnlock()
		return
	}
	lm.RUnlock()
	if !ok {
		fatalServerErrorf("Process %v does not exist in failure detector\n", pid)
	}
	lm.Lock()
	lm.manager[pid] = nil
	lm.Unlock()
}

// Marks a process as up in lm and registers its link object
func (lm *linkManager) markAsUp(pid c.ProcessID, handler *link) {
	lm.RLock()
	link, ok := lm.manager[pid]
	if link != nil && pid != myPhysID {
		fatalServerErrorf("Link to %v already established!\n", pid)
	}
	lm.RUnlock()
	if !ok {
		fatalServerErrorf("Process %v does not exist in failure detector\n", pid)
	}
	lm.Lock()
	lm.manager[pid] = handler
	lm.Unlock()
}

// Returns true iff process pid is up in lm
func (lm *linkManager) isUp(pid c.ProcessID) bool {
	lm.RLock()
	link, ok := lm.manager[pid]
	if !ok {
		fatalServerErrorf("Process %v does not exist in failure detector\n", pid)
	}
	defer lm.RUnlock()
	return link != nil
}

// Returns a slice containing the list of up processes in lm
func (lm *linkManager) getAllUp() []c.ProcessID {
	// The comparator function for sorting a slice of processIDs
	result := make([]c.ProcessID, 0)
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
func (lm *linkManager) getAllDown() []c.ProcessID {
	result := make([]c.ProcessID, 0)
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
	checkFatalServerErrorf(err, "Can't send msg '%v' to master: %v\n", msg, err)
}

// Dials for new connections to all pid < my pid
func (lm *linkManager) dialForConnections() {
	debugPrintf("Dialing for peer connections\n")
	for shouldRun {
		down := linkMgr.getAllDown()
		for _, pid := range down {
			if pid < myPhysID && !linkMgr.isUp(pid) {
				dialingAddr := fmt.Sprintf("%s:%d", gridIP, uint16(basePort)+uint16(pid))
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
	debugPrintf("Listening for peer connections\n")
	listenerAddr := fmt.Sprintf("%s:%d", gridIP, uint16(basePort)+uint16(myPhysID))
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
	debugPrintf("Listening for master connecting on %v\n", masterAddr)
	mstrListener, err := net.Listen("tcp", masterAddr)
	checkFatalServerErrorf(err, "Error connecting to master\n")
	mstrConn, err := mstrListener.Accept()
	checkFatalServerErrorf(err, "Error connecting to master\n")
	lm.masterConn = mstrConn
	debugPrintf("Accepted master connection\n")
	// main loop: process commands from master as async goroutine
	go func() {
		defer lm.masterConn.Close()
		connReader := bufio.NewReader(lm.masterConn)
		for shouldRun {
			data, err := connReader.ReadString('\n')
			checkFatalServerErrorf(err, "Broken connection from master\n")
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
