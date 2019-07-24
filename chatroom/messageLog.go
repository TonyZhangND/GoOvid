package main

// This file contains the definition and methods of the messageLog object.
// A messageLog is an in-memory, thread-safe slice that stores all
// inter-server messages the Ovid server receives, as per the chatroom spec.

import "sync"

type messageLog struct {
	log []string
	sync.RWMutex
}

// Contructor for messageLog
func newMessageLog() *messageLog {
	return &messageLog{log: make([]string, 0)}
}

// Appends message m to the log of ml
func (ml *messageLog) appendMsg(m string) {
	ml.Lock()
	ml.log = append(ml.log, m)
	ml.Unlock()
}

// Returns a copy of the log of ml
func (ml *messageLog) getMessages() []string {
	result := make([]string, len(ml.log))
	ml.RLock()
	for i, m := range ml.log {
		result[i] = m
	}
	ml.RUnlock()
	return result
}
