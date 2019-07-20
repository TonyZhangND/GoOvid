package server

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

// A link l is an object that manages a connection between this process
// and another process p
type link struct {
	conn net.Conn
	// TODO: use an enumerated type for other?
	other    int // who's on the other end of the line. -1 if unknown
	isActive bool
	gotPing  bool // did I receive a ping in the last pingInterval?
	sync.Mutex
}

// Constructor for link where other party is unknown
func newLink(c net.Conn) *link {
	l := &link{conn: c, other: -1, isActive: true, gotPing: false}
	return l
}

// Constructor for link where other party is known
func newLinkKnownOther(c net.Conn, pid processID) *link {
	l := &link{conn: c, other: int(pid), isActive: true, gotPing: false}
	linkMgr.markAsUp(pid, l)
	return l
}

// Close this link. There are a few things to take care of
// 1. Mark myself as inactive to terminate all my infinite loops
// 2. Mark my connection as down
// 3. Close my net.Conn lannel
func (l *link) close() {
	l.isActive = false
	if l.other < 0 {
		return
	}
	linkMgr.markAsDown(processID(l.other))
	l.conn.Close()
}

// Sends the raw string s into l.conn channel
// Note: this method is NOT responsible for formatting the string s
func (l *link) send(s string) {
	_, err := l.conn.Write([]byte(string(s)))
	if err != nil {
		debugPrintln(
			fmt.Sprintf(
				"Send %v from %v to %v failed. Closing connection\n",
				s, myPhysID, l.other))
		l.close()
	}
}

// Begins sending pings into l.conn channel
func (l *link) runPinger() {
	for l.isActive {
		ping := fmt.Sprintf("ping %v\n", myPhysID)
		l.send(ping)
		time.Sleep(pingInterval)
	}
}

// Failure detector: If I did not receive a ping in the last pingInterval, then
// shut down the connection
func (l *link) runCheckState() {
	time.Sleep(pingInterval * 2) // initial grace period
	for l.isActive {
		if !l.gotPing {
			l.close()
			return
		}
		l.gotPing = false
		time.Sleep(pingInterval * 2)
	}
}

// Processes a ping received from the net.Conn channel
func (l *link) doRcvPing(s string) {
	l.gotPing = true
	if l.other < 0 {
		sender, err := strconv.Atoi(s)
		if err != nil {
			errMsg := fmt.Sprintf("Invalid ping %v", s)
			fatalError(errMsg)
		}
		l.other = sender
		linkMgr.markAsUp(processID(sender), l)
	}
}

// Processes a message received from the net.Conn channel
func (l *link) doRcvMsg(s string) {
	sSlice := strings.SplitN(strings.TrimSpace(s), " ", 2)
	// sender := sSlice[0]
	msg := sSlice[1]
	msgLog.appendMsg(msg)
}

// Main thread for eal server connection
func (l *link) handleConnection() {
	defer l.close()
	go l.runPinger()
	go l.runCheckState()
	debugPrintln(fmt.Sprintf("Serving %s", l.conn.RemoteAddr().String()))
	connReader := bufio.NewReader(l.conn)
	for l.isActive {
		data, err := connReader.ReadString('\n')
		if err != nil {
			// the connection is dead. Kill this link
			debugPrintln(
				fmt.Sprintf(
					"Process %v lost connection with %v",
					myPhysID, l.other))
			l.close()
			return
		}
		dataSlice := strings.SplitN(strings.TrimSpace(data), " ", 2)
		header := dataSlice[0]
		payload := dataSlice[1]
		switch header {
		case "ping":
			l.doRcvPing(payload)
		case "msg":
			l.doRcvMsg(payload)
		default:
			debugPrintln(fmt.Sprintf("Invalid msg %v from master\n", header))
		}
	}
}
