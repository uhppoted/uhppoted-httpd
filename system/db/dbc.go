package db

import (
	"sync"

	"github.com/uhppoted/uhppoted-httpd/audit"
)

type DBC interface {
	Write(audit.AuditRecord)
	Commit()
}

type dbc struct {
	trail audit.AuditTrail
	logs  []audit.AuditRecord
	guard sync.Mutex
}

func NewDBC(trail audit.AuditTrail) DBC {
	return &dbc{
		trail: trail,
		logs:  []audit.AuditRecord{},
	}
}

func (d *dbc) Write(record audit.AuditRecord) {
	d.guard.Lock()
	defer d.guard.Unlock()

	d.logs = append(d.logs, record)
}

func (d *dbc) Commit() {
	d.guard.Lock()
	defer d.guard.Unlock()

	if d.trail != nil {
		for _, r := range d.logs {
			d.trail.Write(r)
		}
	}

	d.logs = []audit.AuditRecord{}
}
