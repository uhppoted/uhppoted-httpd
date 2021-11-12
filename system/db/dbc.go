package db

import (
	"sync"

	"github.com/uhppoted/uhppoted-httpd/audit"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
)

type DBC interface {
	Stash([]catalog.Object)
	Objects() []catalog.Object
	Commit()
	Write(audit.AuditRecord)
}

type dbc struct {
	objects []catalog.Object
	trail   audit.AuditTrail
	logs    []audit.AuditRecord
	guard   sync.Mutex
}

func NewDBC(trail audit.AuditTrail) DBC {
	return &dbc{
		objects: []catalog.Object{},
		trail:   trail,
		logs:    []audit.AuditRecord{},
	}
}

func (d *dbc) Stash(list []catalog.Object) {
	d.guard.Lock()
	defer d.guard.Unlock()

	if list != nil {
		d.objects = append(d.objects, list...)
	}
}

func (d *dbc) Objects() []catalog.Object {
	return d.objects
}

func (d *dbc) Commit() {
	d.guard.Lock()
	defer d.guard.Unlock()

	if d.objects != nil {
		for _, o := range d.objects {
			catalog.Put(o.OID, o.Value)
		}
	}

	if d.trail != nil {
		for _, r := range d.logs {
			d.trail.Write(r)
		}
	}

	d.logs = []audit.AuditRecord{}
}

func (d *dbc) Write(record audit.AuditRecord) {
	d.guard.Lock()
	defer d.guard.Unlock()

	d.logs = append(d.logs, record)
}
