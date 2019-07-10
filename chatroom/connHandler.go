package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// A connHandler ch is an object that manages a connection between this process
// and another process p
type connHandler struct {
	conn net.Conn
	// TODO: use an enumerated type for other?
	other    int // who's on the other end of the line. -1 if unknown
	isActive bool
}

// Constructor for connHander where other party is unknown
func newConnHandler(c net.Conn) *connHandler {
	ch := &connHandler{conn: c, other: -1, isActive: true}
	return ch
}

// Constructor for connHander where other party is known
func newConnHandlerKnownOther(c net.Conn, pid processID) *connHandler {
	ch := &connHandler{conn: c, other: int(pid), isActive: true}
	connRouter.markAsUp(pid, ch)
	return ch
}

// Close this connection. There are a few things to take care of
// 1. Mark myself as inactive to terminate all my infinite loops
// 2. Mark my connection as down
// 3. Close my net.Conn channel
func (ch *connHandler) close() {
	ch.isActive = false
	if ch.other < 0 {
		return
	}
	connRouter.markAsDown(processID(ch.other))
	err := ch.conn.Close()
	if err != nil {
		fmt.Printf("Error closing connection by %v\n", myPhysID)
		os.Exit(1)
	}
}

// Sends the raw string s into ch.conn channel
// Note: this method is NOT responsible for formatting the string s
func (ch *connHandler) send(s string) {
	_, err := ch.conn.Write([]byte(string(s)))
	if err != nil {
		fmt.Printf("Error sending msg %v. Closing connection\n", s)
		ch.close()
	}
}

// Begins sending pings into ch.conn channel
func (ch *connHandler) beginPinging() {
	for ch.isActive {
		ping := fmt.Sprintf("ping %v\n", myPhysID)
		ch.send(ping)
		// fmt.Printf("%v sending ping to %v\n", myPhysID, ch.other)
		time.Sleep(1 * time.Second)
	}
}

// Processes a ping received from the net.Conn channel
func (ch *connHandler) doRcvPing(s string) {
	sender, err := strconv.Atoi(s)
	if err != nil {
		fmt.Printf("Error, Invalid ping received by %v\n", myPhysID)
		os.Exit(1)
	}
	ch.other = sender
	connRouter.markAsUp(processID(sender), ch)
}

// Processes a message received from the net.Conn channel
func (ch *connHandler) doRcvMsg(s string) {
	// sSlice := strings.SplitN(strings.TrimSpace(s), " ", 2)
	// sender := sSlice[0]
	// msg := sSlice[1]
}

// Main thread for each server connection
func (ch *connHandler) handleConnection() {
	// TODO: Spawn a failure detector
	defer ch.close()
	go ch.beginPinging()
	fmt.Printf("Serving %s\n", ch.conn.RemoteAddr().String())
	for ch.isActive {
		data, err := bufio.NewReader(ch.conn).ReadString('\n')
		if err != nil {
			fmt.Println(err)
		}
		// fmt.Printf("%v received %v", myPhysID, data)
		dataSlice := strings.SplitN(strings.TrimSpace(data), " ", 2)
		header := dataSlice[0]
		payload := dataSlice[1]
		switch header {
		case "ping":
			ch.doRcvPing(payload)
		case "msg":
			ch.doRcvMsg(payload)
		default:
			fmt.Printf("Error, Invalid msg %v from master\n", header)
		}
		time.Sleep(2 * time.Second)
	}
}
