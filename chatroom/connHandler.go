package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

type connHandler struct {
	conn  net.Conn
	other int // who's on the other end of the line
}

func newConnHandler(c net.Conn) *connHandler {
	return &connHandler{conn: c, other: -1}
}

func newConnHandlerKnownOther(c net.Conn, pid processID) *connHandler {
	ch := &connHandler{conn: c, other: int(pid)}
	connRouter.markAsUp(pid, ch)
	return ch
}

func (ch *connHandler) close() {
	if ch.other < 0 {
		return
	}
	err := ch.conn.Close()
	if err != nil {
		fmt.Printf("Error closing connection by %v\n", myPhysID)
		os.Exit(1)
	}
	connRouter.markAsDown(processID(ch.other))
}

func (ch *connHandler) send(s string) {
	_, err := ch.conn.Write([]byte(string(s)))
	if err != nil {
		fmt.Printf("Error sending msg %v. Closing connection\n", s)
		ch.conn.Close()
	}
}

// TODO: Main thread for each server connection
func (ch *connHandler) handleConnection(conn net.Conn) {
	// TODO: Spawn a failure detector
	defer conn.Close()
	fmt.Printf("Serving %s\n", conn.RemoteAddr().String())
	for shouldRun {
		data, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Println(err)
		}
		dataSlice := strings.SplitN(strings.TrimSpace(data), " ", 2)
		header := dataSlice[0]
		payload := dataSlice[1]
		fmt.Printf("Type of message: %v\n", header)
		switch header {
		case "ping":
			doRcvPing(payload, conn)
		case "msg":
			doRcvMsg(payload)
		default:
			fmt.Printf("Error, Invalid msg %v from master\n", header)
		}
		fmt.Println("Done responding to master")
		time.Sleep(2 * time.Second)
	}
}

func doRcvPing(s string, conn net.Conn) {
	// sender, err := strconv.Atoi(s)
	// if err != nil {
	// 	fmt.Printf("Error, Invalid ping received by %v\n", myPhysID)
	// 	os.Exit(1)
	// }
	// // check that sender is a valid process
	// if !connRouter.isKnown(processID(sender)) {
	// 	fmt.Printf("Error, Invalid ping from  %v received by %v\n",
	// 		myPhysID, sender)
	// 	os.Exit(1)
	// }
	// connRouter.markAsUp(processID(sender), conn)
}

func doRcvMsg(s string) {
	// sSlice := strings.SplitN(strings.TrimSpace(s), " ", 2)
	// sender := sSlice[0]
	// msg := sSlice[1]
}
