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

	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	"github.com/uhppoted/uhppoted-lib/eventlog"
)

type AuditTrail interface {
	Write(record ...AuditRecord)
}

type trail struct {
	logger *log.Logger
}

type Details struct {
	ID          string
	Name        string
	Field       string
	Description string
	Before      string
	After       string
}

type AuditRecord struct {
	Timestamp time.Time
	UID       string
	OID       schema.OID
	Component string
	Operation string
	Details   Details
}

var auditTrail = trail{}

var guard sync.Mutex

func MakeTrail() AuditTrail {
	return &auditTrail
}

func SetAuditFile(file string) error {
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

	// Basic sanity check because log.Printf(...) does not return an error if the logfile
	// is not writeable
	return logger.Output(2, "AUDIT TRAIL START")
}

func (t *trail) Write(records ...AuditRecord) {
	for _, record := range records {
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
	}
}
