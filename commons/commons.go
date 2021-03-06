package commons

import (
	"fmt"
	"net"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
)

type (
	// ProcessID is a type representing the physical and virtual ID's of an agent
	ProcessID uint16
	// PortNum is a type representing an IP port on a host
	PortNum uint16
	// BoxID is the unique, canonical address representing each box
	BoxID string
)

// Route is a tuple struct representing a route
type Route struct {
	DestID   ProcessID
	DestPort PortNum
}

// ParseBoxAddr parses string s into a canonical box address
func ParseBoxAddr(s string) BoxID {
	ipStr, portStr, err := net.SplitHostPort(s)
	CheckFatalOvidErrorf(err, "Cannot parse box string %s\n", s)
	ip := net.ParseIP(ipStr)
	if ip == nil {
		FatalOvidErrorf("Cannot parse IP %s of box %s\n", ipStr, s)
	}
	port, err := strconv.ParseUint(portStr, 10, 16)
	CheckFatalOvidErrorf(err, "Cannot parse port %s of box %s\n", portStr, s)
	if strings.Contains(ip.String(), ":") {
		// IPv6
		return BoxID(fmt.Sprintf("[%s]:%d", ip.String(), port))
	}
	// IPv4
	return BoxID(fmt.Sprintf("%s:%d", ip.String(), port))
}

// FatalOvidErrorf prints the error messange and kills the entire program
func FatalOvidErrorf(s string, a ...interface{}) {
	errMsg := fmt.Sprintf(s, a...)
	fmt.Printf("Error : Ovid : %s", errMsg)
	debug.PrintStack()
	os.Exit(1)
}

// CheckFatalOvidErrorf prints the error messange and kills the entire program
// if an error is detected
func CheckFatalOvidErrorf(err error, s string, a ...interface{}) {
	if err != nil {
		FatalOvidErrorf(s, a...)
		os.Exit(1)
	}
}
