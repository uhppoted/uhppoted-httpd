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

	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-lib/eventlog"
)

type trail struct {
	logger    *log.Logger
	listeners []chan<- AuditRecord
}

type Details struct {
	ID          string
	Name        string
	Field       string
	Description string
}

type AuditRecord struct {
	Timestamp time.Time
	UID       string
	OID       catalog.OID
	Component string
	Operation string
	Details   Details
}

var auditTrail = trail{
	listeners: []chan<- AuditRecord{},
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

func AddListener(listener chan<- AuditRecord) {
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

func Write(record AuditRecord) {
	auditTrail.Write(record)
}

func (t *trail) Write(record AuditRecord) {
	var logmsg string
	if info, err := json.Marshal(record.Details); err == nil {
		logmsg = fmt.Sprintf("%-10v %-10v %-10v %s", record.UID, record.Component, record.Operation, info)
	} else {
		logmsg = fmt.Sprintf("%-10v %-10v %-10v %v", record.UID, record.Component, record.Operation, record.Details)
	}

	if t.logger != nil {
		t.logger.Printf("%s", logmsg)
	} else {
		log.Printf("%s", logmsg)
	}

	for _, l := range t.listeners {
		l <- record
	}
}
