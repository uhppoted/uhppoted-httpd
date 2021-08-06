package audit

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/uhppoted/uhppoted-lib/eventlog"
)

type Trail interface {
	Write(LogEntry)
}

type trail struct {
	logger *log.Logger
}

type LogEntry struct {
	UID       string
	Module    string
	Operation string
	Info      interface{}
}

func NewAuditTrail(file string) *trail {
	events := eventlog.Ticker{Filename: file, MaxSize: 10}
	logger := log.New(&events, "", log.Ldate|log.Ltime|log.LUTC)
	rotate := make(chan os.Signal, 1)

	signal.Notify(rotate, syscall.SIGHUP)

	go func() {
		for {
			<-rotate
			log.Printf("Rotating audit trail file '%s'\n", file)
			events.Rotate()
		}
	}()

	return &trail{
		logger: logger,
	}
}

func (l *trail) Write(entry LogEntry) {
	var logmsg string
	if info, err := json.Marshal(entry.Info); err == nil {
		logmsg = fmt.Sprintf("%-10v %-10v %-10v %s", entry.UID, entry.Module, entry.Operation, info)
	} else {
		logmsg = fmt.Sprintf("%-10v %-10v %-10v %v", entry.UID, entry.Module, entry.Operation, entry.Info)
	}

	if l.logger != nil {
		l.logger.Printf("%s", logmsg)
	} else {
		log.Printf("%s", logmsg)
	}
}
