package db

import (
	"sync"

	"github.com/uhppoted/uhppoted-httpd/audit"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
)

type DBC interface {
	Stash([]schema.Object)
	Updated(oid schema.OID, suffix schema.Suffix, value any)
	Objects() []schema.Object
	Commit(sys System)
	Write(audit.AuditRecord)
}

type System interface {
	Update(oid schema.OID, field schema.Suffix, value any)
}

type dbc struct {
	objects []schema.Object
	updated []update
	trail   audit.AuditTrail
	logs    []audit.AuditRecord
	sync.Mutex
}

type update struct {
	object schema.OID
	field  schema.Suffix
	value  any
}

func NewDBC(trail audit.AuditTrail) DBC {
	return &dbc{
		objects: []schema.Object{},
		updated: []update{},
		trail:   trail,
		logs:    []audit.AuditRecord{},
	}
}

func (d *dbc) Stash(list []schema.Object) {
	d.Lock()
	defer d.Unlock()

	if list != nil {
		d.objects = append(d.objects, list...)
	}
}

func (d *dbc) Updated(oid schema.OID, field schema.Suffix, value any) {
	d.Lock()
	defer d.Unlock()

	d.updated = append(d.updated, update{
		object: oid,
		field:  field,
		value:  value,
	})
}

func (d *dbc) Objects() []schema.Object {
	return d.objects
}

func (d *dbc) Commit(sys System) {
	d.Lock()
	defer d.Unlock()

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

	for _, v := range d.updated {
		sys.Update(v.object, v.field, v.value)
	}
}

func (d *dbc) Write(record audit.AuditRecord) {
	d.Lock()
	defer d.Unlock()

	d.logs = append(d.logs, record)
}
