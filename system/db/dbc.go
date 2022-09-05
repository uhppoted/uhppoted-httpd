package db

import (
	"fmt"

	"github.com/uhppoted/uhppoted-httpd/audit"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
)

type DBC struct {
	impl impl
}

type impl interface {
	Stash([]schema.Object)
	Updated(oid schema.OID, suffix schema.Suffix, value any)
	Objects() []schema.Object
	Commit(sys System, hook func())
	Log(uid, operation string, OID schema.OID, component string, ID, name any, field string, before, after any, format string, fields ...any)
}

type System interface {
	Update(oid schema.OID, field schema.Suffix, value any)
}

func NewDBC(trail audit.AuditTrail) DBC {
	return DBC{
		impl: &dbc{
			objects: []schema.Object{},
			updated: []update{},
			trail:   trail,
			logs:    []audit.AuditRecord{},
		},
	}
}

func (d *DBC) Stash(list []schema.Object) {
	if d != nil && d.impl != nil {
		d.impl.Stash(list)
	}
}

func (d *DBC) Updated(oid schema.OID, field schema.Suffix, value any) {
	if d != nil && d.impl != nil {
		d.impl.Updated(oid, field, value)
	}
}

func (d *DBC) Objects() []schema.Object {
	if d != nil && d.impl != nil {
		return d.impl.Objects()
	}

	return []schema.Object{}
}

func (d *DBC) Commit(system System, hook func()) {
	if d != nil && d.impl != nil {
		d.impl.Commit(system, hook)
	}
}

func (d *DBC) Log(uid, operation string, OID schema.OID, component string, ID, name any, field string, before, after any, format string, fields ...any) {
	if d != nil && d.impl != nil {
		d.impl.Log(uid, operation, OID, component, ID, name, field, before, after, format, fields...)
	}
}

// Returns a deduplicated list of objects, retaining only the the last (i.e. latest) value.
// NOTE: this implementation is horrifically inefficient but the list is expected to almost
//
//	always be tiny since it is the result of a manual edit.
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
