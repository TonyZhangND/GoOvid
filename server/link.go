package server

// This file contains the definition and methods of the link object.
// A link is a wrapper for the TCP connection between two servers.
// It is responsible for maintaining and monitoring the health of the
// connection.

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"

	c "github.com/TonyZhangND/GoOvid/commons"
)

// A link l is an object that manages a connection between this process
// and another process p
type link struct {
	conn          net.Conn
	other         c.BoxID     // who's on the other end of the line. "" if unknown
	isActive      bool        // loop condition for the link's routines
	serverOutChan chan string // used to stream messages to main server loop
}

// Constructor for link where other party is unknown
func newLink(c net.Conn, sOutChan chan string) *link {
	l := &link{conn: c, other: "", isActive: false, serverOutChan: sOutChan}
	return l
}

// Constructor for link where other party is known
func newLinkKnownOther(c net.Conn, bid c.BoxID, sOutChan chan string) *link {
	l := &link{conn: c, other: bid, isActive: true, serverOutChan: sOutChan}
	linkMgr.markAsUp(bid, l)
	return l
}

// Close this link. There are a few things to take care of
// 1. Mark myself as inactive to terminate all my infinite loops
// 2. Mark my connection as down
// 3. Close my net.Conn channel
func (l *link) close() {
	l.isActive = false
	if string(l.other) == "" {
		return
	}
	linkMgr.markAsDown(c.BoxID(l.other))
	l.conn.Close()
}

// Sends the raw string s into l.conn channel
// Note: this method is NOT responsible for formatting the string s
func (l *link) send(s string) {
	_, err := l.conn.Write([]byte(string(s)))
	if err != nil {
		debugPrintf("Send %v to %v failed. Closing connection\n", s, l.other)
		l.close()
	}
}

// Begins sending pings into l.conn channel
func (l *link) runPinger() {
	for l.isActive {
		ping := fmt.Sprintf("ping %v\n", myBoxID)
		l.send(ping)
		time.Sleep(pingInterval)
	}
}

// Processes a ping received from the net.Conn channel
func (l *link) doRcvPing(s string) {
	if string(l.other) == "" {
		sender := c.ParseBoxAddr(s)
		l.other = sender
		linkMgr.markAsUp(sender, l)
	}
}

// Main thread for server-server connection
func (l *link) handleConnection() {
	defer l.close()
	l.isActive = true
	go l.runPinger()
	debugPrintf("Serving %s\n", l.conn.RemoteAddr().String())
	connReader := bufio.NewReader(l.conn)
	inChan := make(chan string)
	go func() {
		for l.isActive {
			data, err := connReader.ReadString('\n')
			if err != nil {
				// the connection is dead. Kill this link
				debugPrintf("Lost connection with %v\n", l.other)
				l.close()
				return
			}
			inChan <- data
		}
	}()
	for l.isActive {
		select {
		case data := <-inChan:
			dataSlice := strings.SplitN(strings.TrimSpace(data), " ", 2)
			header := dataSlice[0]
			payload := dataSlice[1]
			switch header {
			case "ping":
				l.doRcvPing(payload)
			case "msg":
				l.serverOutChan <- payload
			default:
				debugPrintf("Invalid msg %v\n", header)
			}
		case <-time.After(pingInterval * 2):
			l.close()
			return
		}
	}
}
