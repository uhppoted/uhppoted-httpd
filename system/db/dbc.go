package db

import (
	"sync"

	"github.com/uhppoted/uhppoted-httpd/audit"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type DBC interface {
	Stash([]schema.Object)
	Updated(controller types.IController, suffix schema.Suffix, value any)
	Objects() []schema.Object
	Commit(sys System)
	Write(audit.AuditRecord)
}

type System interface {
	Update(controller types.IController, field schema.Suffix, value any)
}

type dbc struct {
	objects []schema.Object
	updated []update
	trail   audit.AuditTrail
	logs    []audit.AuditRecord
	sync.Mutex
}

type update struct {
	controller types.IController
	field      schema.Suffix
	value      any
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

func (d *dbc) Updated(controller types.IController, field schema.Suffix, value any) {
	d.Lock()
	defer d.Unlock()

	d.updated = append(d.updated, update{
		controller: controller,
		field:      field,
		value:      value,
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
		sys.Update(v.controller, v.field, v.value)
	}
}

func (d *dbc) Write(record audit.AuditRecord) {
	d.Lock()
	defer d.Unlock()

	d.logs = append(d.logs, record)
}
