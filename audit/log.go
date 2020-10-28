package audit

import (
	"encoding/json"
	"log"
)

type Log struct {
}

type LogEntry struct {
	UID       string
	Module    string
	Operation string
	Info      interface{}
}

func NewAuditTrail() Log {
	return Log{}
}

func (l *Log) Write(entry LogEntry) {
	if info, err := json.Marshal(entry.Info); err == nil {
		log.Printf("%-10v %-10v %-10v %s", entry.UID, entry.Module, entry.Operation, info)
	} else {
		log.Printf("%-10v %-10v %-10v %+v", entry.UID, entry.Module, entry.Operation, entry.Info)
	}
}
