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

type connHandler struct {
	conn     net.Conn
	other    int // who's on the other end of the line
	isActive bool
}

func newConnHandler(c net.Conn) *connHandler {
	ch := &connHandler{conn: c, other: -1, isActive: true}
	go ch.beginPinging()
	return ch
}

func newConnHandlerKnownOther(c net.Conn, pid processID) *connHandler {
	ch := &connHandler{conn: c, other: int(pid), isActive: true}
	connRouter.markAsUp(pid, ch)
	go ch.beginPinging()
	return ch
}

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

func (ch *connHandler) send(s string) {
	_, err := ch.conn.Write([]byte(string(s)))
	if err != nil {
		fmt.Printf("Error sending msg %v. Closing connection\n", s)
		ch.conn.Close()
	}
}

func (ch *connHandler) beginPinging() {
	for ch.isActive {
		ping := fmt.Sprintf("ping %v\n", myPhysID)
		ch.send(ping)
		// fmt.Printf("%v sending ping to %v\n", myPhysID, ch.other)
		time.Sleep(1 * time.Second)
	}
}

func (ch *connHandler) doRcvPing(s string, conn net.Conn) {
	sender, err := strconv.Atoi(s)
	if err != nil {
		fmt.Printf("Error, Invalid ping received by %v\n", myPhysID)
		os.Exit(1)
	}
	ch.other = sender
	connRouter.markAsUp(processID(sender), ch)
}

func (ch *connHandler) doRcvMsg(s string) {
	// sSlice := strings.SplitN(strings.TrimSpace(s), " ", 2)
	// sender := sSlice[0]
	// msg := sSlice[1]
}

// TODO: Main thread for each server connection
func (ch *connHandler) handleConnection() {
	// TODO: Spawn a failure detector
	defer ch.conn.Close()
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
			ch.doRcvPing(payload, ch.conn)
		case "msg":
			ch.doRcvMsg(payload)
		default:
			fmt.Printf("Error, Invalid msg %v from master\n", header)
		}
		time.Sleep(2 * time.Second)
	}
}
