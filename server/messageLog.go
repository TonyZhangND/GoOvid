package server

import "sync"

type messageLog struct {
	log []string
	sync.RWMutex
}

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
