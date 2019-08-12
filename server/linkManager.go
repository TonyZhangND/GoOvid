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
	manager       map[c.BoxID]*link
	masterConn    net.Conn    // connection with the master program
	serverOutChan chan string //used to stream inter-server messages to main server loop
	masterOutChan chan string // used to stream master messages to main server loop
	sync.RWMutex
}

// Constructor for linkManager
// It takes a slice of all known box IDs, and initializes a
// connTracker lm with lm[p]=nil for all p in knownBoxes.
func newLinkManager(knownBoxes []c.BoxID,
	sOutChan chan string,
	mstrOutChan chan string) *linkManager {
	t := make(map[c.BoxID]*link)
	for _, bid := range knownBoxes {
		t[bid] = nil
	}
	return &linkManager{
		manager:       t,
		serverOutChan: sOutChan,
		masterOutChan: mstrOutChan}
}

// Marks a box as down in lm and de-registers its link object
func (lm *linkManager) markAsDown(bid c.BoxID) {
	lm.RLock()
	link, ok := lm.manager[bid]
	if link == nil {
		defer lm.RUnlock()
		return
	}
	lm.RUnlock()
	if !ok {
		fatalServerErrorf("Process %v does not exist in failure detector\n", bid)
	}
	lm.Lock()
	lm.manager[bid] = nil
	lm.Unlock()
}

// Marks a box as up in lm and registers its link object
func (lm *linkManager) markAsUp(bid c.BoxID, handler *link) {
	lm.RLock()
	link, ok := lm.manager[bid]
	if link != nil && bid != myBoxID {
		fatalServerErrorf("Link to %v already established!\n", bid)
	}
	lm.RUnlock()
	if !ok {
		fatalServerErrorf("Process %v does not exist in failure detector\n", bid)
	}
	lm.Lock()
	lm.manager[bid] = handler
	lm.Unlock()
}

// Returns true iff box bid is up in lm
func (lm *linkManager) isUp(bid c.BoxID) bool {
	lm.RLock()
	link, ok := lm.manager[bid]
	if !ok {
		fatalServerErrorf("Process %v does not exist in failure detector\n", bid)
	}
	defer lm.RUnlock()
	return link != nil
}

// Returns a slice containing the list of up boxes in lm
func (lm *linkManager) getAllUp() []c.BoxID {
	// The comparator function for sorting a slice of processIDs
	result := make([]c.BoxID, 0)
	lm.RLock()
	for bid, link := range lm.manager {
		// find the nodes that are up
		if link != nil {
			result = append(result, bid)
		}
	}
	result = append(result, myBoxID) // Cogito ergo sum
	defer lm.RUnlock()
	return result
}

// Returns a slice containing the list of down boxes in lm
func (lm *linkManager) getAllDown() []c.BoxID {
	result := make([]c.BoxID, 0)
	lm.RLock()
	for pid, link := range lm.manager {
		if link == nil && pid != myBoxID { // cogito ergo sum
			result = append(result, pid)
		}
	}
	defer lm.RUnlock()
	return result
}

// Returns a slice containing the list of boxes in lm
func (lm *linkManager) getAllKnown() []c.BoxID {
	lm.RLock()
	result := make([]c.BoxID, len(lm.manager))
	i := 0
	for bid := range lm.manager {
		result[i] = bid
		i++
	}
	defer lm.RUnlock()
	return result
}

// Sends msg on all channels.
// Applies Ovid message format and headers
func (lm *linkManager) broadcast(msg string) {
	s := fmt.Sprintf("chatroom %v %s\n", myBoxID, msg)
	lm.serverOutChan <- s // first send to myself
	lm.RLock()
	for _, link := range lm.manager {
		if link != nil {
			link.send(s)
		}
	}
	lm.RUnlock()
}

// Sends msg to destBox, given that destBox is up
// Applies Ovid message format and headers
func (lm *linkManager) send(destBox c.BoxID, msg string) {
	if destBox == myBoxID {
		c.FatalOvidErrorf("Intra-server messages should not reach linkManager layer\n")
	} else {
		// Sending to other box
		lm.RLock() // We lock so that the link won't be pulled from beneath out feet
		if lm.isUp(destBox) {
			s := fmt.Sprintf("msg %s\n", msg)
			lm.manager[destBox].send(s)
		}
		lm.RLock()
	}
}

// Sends msg string to the master
func (lm *linkManager) sendToMaster(msg string) {
	_, err := lm.masterConn.Write([]byte(msg + "\n"))
	checkFatalServerErrorf(err, "Can't send msg '%v' to master: %v\n", msg, err)
}

// Dials for new connections to all bid < my bid, using build-in string comp
func (lm *linkManager) dialForConnections() {
	debugPrintf("Dialing for peer connections\n")
	for shouldRun {
		down := linkMgr.getAllDown()
		for _, bid := range down {
			if bid < myBoxID && !linkMgr.isUp(bid) {
				// conviniently, the bid is the tcp addr
				c, err := net.DialTimeout("tcp", string(bid),
					20*time.Millisecond)
				if err == nil {
					l := newLinkKnownOther(c, bid, lm.serverOutChan)
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
	// conviniently, the myBoxID is my tcp addr
	l, err := net.Listen("tcp", string(myBoxID))
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
	if masterPort > 0 {
		// // only start master conn if port specified
		go lm.connectAndHandleMaster()
	}
}
