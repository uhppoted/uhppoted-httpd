package db

import (
	"sync"

	"github.com/uhppoted/uhppoted-httpd/audit"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
)

type DBC interface {
	Stash([]schema.Object)
	Objects() []schema.Object
	Commit()
	Write(audit.AuditRecord)
}

type dbc struct {
	objects []schema.Object
	trail   audit.AuditTrail
	logs    []audit.AuditRecord
	guard   sync.Mutex
}

func NewDBC(trail audit.AuditTrail) DBC {
	return &dbc{
		objects: []schema.Object{},
		trail:   trail,
		logs:    []audit.AuditRecord{},
	}
}

func (d *dbc) Stash(list []schema.Object) {
	d.guard.Lock()
	defer d.guard.Unlock()

	if list != nil {
		d.objects = append(d.objects, list...)
	}
}

func (d *dbc) Objects() []schema.Object {
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
