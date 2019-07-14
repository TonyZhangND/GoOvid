package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
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
	sync.Mutex
}

// Constructor for link where other party is unknown
func newLink(c net.Conn) *link {
	l := &link{conn: c, other: -1, isActive: true}
	return l
}

// Constructor for link where other party is known
func newLinkKnownOther(c net.Conn, pid processID) *link {
	l := &link{conn: c, other: int(pid), isActive: true}
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
	err := l.conn.Close()
	if err != nil {
		fmt.Printf("Error closing connection by %v\n", myPhysID)
		os.Exit(1)
	}
}

// Sends the raw string s into l.conn channel
// Note: this method is NOT responsible for formatting the string s
func (l *link) send(s string) {
	_, err := l.conn.Write([]byte(string(s)))
	if err != nil {
		fmt.Printf("Error sending msg %v. Closing connection\n", s)
		l.close()
	}
}

// Begins sending pings into l.conn channel
func (l *link) beginPinging() {
	for l.isActive {
		ping := fmt.Sprintf("ping %v\n", myPhysID)
		l.send(ping)
		time.Sleep(1 * time.Second)
	}
}

// Processes a ping received from the net.Conn channel
func (l *link) doRcvPing(s string) {
	sender, err := strconv.Atoi(s)
	if err != nil {
		fmt.Printf("Error, Invalid ping received by %v\n", myPhysID)
		os.Exit(1)
	}
	l.other = sender
	linkMgr.markAsUp(processID(sender), l)
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
	go l.beginPinging()
	fmt.Printf("Serving %s\n", l.conn.RemoteAddr().String())
	for l.isActive {
		data, err := bufio.NewReader(l.conn).ReadString('\n')
		if err != nil {
			fmt.Println(err)
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
			fmt.Printf("Error, Invalid msg %v from master\n", header)
		}
	}
}
