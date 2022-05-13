package db

import (
	"fmt"
	"sync"

	"github.com/uhppoted/uhppoted-httpd/audit"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
)

type DBC interface {
	Stash([]schema.Object)
	Updated(oid schema.OID, suffix schema.Suffix, value any)
	Objects() []schema.Object
	Commit(sys System, hook func())
	Log(uid, operation string, OID schema.OID, component string, ID, name any, field string, before, after any, format string, fields ...any)
}

type System interface {
	Update(oid schema.OID, field schema.Suffix, value any)
	Updated()
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

// Returns a deduplicated list of objects, retaining only the the last (i.e. latest) value.
func (d *dbc) Objects() []schema.Object {
	return squoosh(d.objects)
}

func (d *dbc) Commit(sys System, hook func()) {
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

	hook()

	for _, v := range d.updated {
		sys.Update(v.object, v.field, v.value)
	}

	sys.Updated()
}

func (d *dbc) Log(uid, operation string, OID schema.OID, component string, ID, name any, field string, before, after any, format string, fields ...any) {
	if d != nil {
		d.Lock()
		defer d.Unlock()

		record := audit.AuditRecord{
			UID:       uid,
			OID:       OID,
			Component: component,
			Operation: operation,
			Details: audit.Details{
				ID:          stringify(ID),
				Name:        stringify(name),
				Field:       field,
				Description: fmt.Sprintf(format, fields...),
				Before:      stringify(before),
				After:       stringify(after),
			},
		}

		d.logs = append(d.logs, record)
	}
}

// Returns a deduplicated list of objects, retaining only the the last (i.e. latest) value.
// NOTE: this implementation is horrifically inefficient but the list is expected to almost
//       always be tiny since it is the result of a manual edit.
func squoosh(objects []schema.Object) []schema.Object {
	keys := map[schema.OID]struct{}{}
	list := []schema.Object{}

	for i := len(objects); i > 0; i-- {
		object := objects[i-1]
		oid := object.OID
		if _, ok := keys[oid]; !ok {
			keys[oid] = struct{}{}
			list = append([]schema.Object{object}, list...)
		}
	}

	return list
}

func stringify(a any) string {
	switch v := a.(type) {
	case *uint32:
		if v != nil {
			return fmt.Sprintf("%v", *v)
		}

	case *string:
		if v != nil {
			return fmt.Sprintf("%v", *v)
		}

	default:
		return fmt.Sprintf("%v", a)
	}

	return ""
}
