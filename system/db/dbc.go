package db

import (
	"sync"

	"github.com/uhppoted/uhppoted-httpd/audit"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
)

type DBC interface {
	Write(audit.AuditRecord)
	Commit([]catalog.Object)
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

func (d *dbc) Commit(objects []catalog.Object) {
	d.guard.Lock()
	defer d.guard.Unlock()

	if objects != nil {
		for _, o := range objects {
			catalog.PutV(o.OID, o.Value, false)
		}
	}

	if d.trail != nil {
		for _, r := range d.logs {
			d.trail.Write(r)
		}
	}

	d.logs = []audit.AuditRecord{}
}
