package audit

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/uhppoted/uhppoted-lib/eventlog"
)

type trail struct {
	logger    *log.Logger
	listeners []chan<- LogEntry
}

type Info interface {
	Field() string
}

type LogEntry struct {
	Timestamp time.Time
	UID       string
	Module    string
	Operation string
	Info      Info
}

var auditTrail = trail{
	listeners: []chan<- LogEntry{},
}

var guard sync.Mutex

func SetAuditFile(file string) {
	guard.Lock()
	defer guard.Unlock()

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

	auditTrail.logger = logger
}

func AddListener(listener chan<- LogEntry) {
	guard.Lock()
	defer guard.Unlock()

	if listener != nil {
		for _, l := range auditTrail.listeners {
			if l == listener {
				return
			}
		}

		auditTrail.listeners = append(auditTrail.listeners, listener)
	}
}

func Write(entry LogEntry) {
	auditTrail.Write(entry)
}

func (t *trail) Write(entry LogEntry) {
	var logmsg string
	if info, err := json.Marshal(entry.Info); err == nil {
		logmsg = fmt.Sprintf("%-10v %-10v %-10v %s", entry.UID, entry.Module, entry.Operation, info)
	} else {
		logmsg = fmt.Sprintf("%-10v %-10v %-10v %v", entry.UID, entry.Module, entry.Operation, entry.Info)
	}

	if t.logger != nil {
		t.logger.Printf("%s", logmsg)
	} else {
		log.Printf("%s", logmsg)
	}

	for _, l := range t.listeners {
		l <- entry
	}
}
