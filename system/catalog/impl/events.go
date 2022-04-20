package memdb

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
)

type events struct {
	base schema.OID
	m    map[eventKey]*event
	last uint32
}

type event struct {
	OID      schema.OID
	deviceID uint32
	index    uint32
	deleted  bool
}

type eventKey struct {
	deviceID uint32
	index    uint32
}

func (t *events) New(v any) schema.OID {
	u := v.(catalog.CatalogEvent)

	key := eventKey{
		deviceID: u.DeviceID,
		index:    u.Index,
	}

	if e, ok := t.m[key]; ok {
		return e.OID
	}

	suffix := t.last + 1
	oid := schema.OID(fmt.Sprintf("%v.%d", t.base, suffix))

	t.m[key] = &event{
		OID:      oid,
		deviceID: u.DeviceID,
		index:    u.Index,
	}

	t.last = suffix

	return oid
}

func (t *events) Put(oid schema.OID, v any) {
	u := v.(catalog.CatalogEvent)

	key := eventKey{
		deviceID: u.DeviceID,
		index:    u.Index,
	}

	if !oid.HasPrefix(t.base) {
		panic(fmt.Sprintf("PUT: illegal oid %v for base %v", oid, t.base))
	}

	if e, ok := t.m[key]; ok && oid == e.OID {
		return
	} else if ok && oid != e.OID {
		panic(fmt.Sprintf("PUT: oid %v for event {%v,%v} does not matched existing OID %v", oid, e.deviceID, e.index, e.OID))
	}

	suffix := strings.TrimPrefix(string(oid), string(t.base))
	match := regexp.MustCompile(`\.([0-9]+)`).FindStringSubmatch(suffix)
	if match == nil || len(match) != 2 {
		panic(fmt.Sprintf("PUT: invalid event oid %v", oid))
	}

	index, err := strconv.ParseUint(match[1], 10, 32)
	if err != nil {
		panic(fmt.Sprintf("PUT: out of range oid %v for base %v", oid, t.base))
	}

	t.m[key] = &event{
		OID:      oid,
		deviceID: u.DeviceID,
		index:    u.Index,
	}

	if v := uint32(index); v > t.last {
		t.last = v
	}
}

// Horrifically inefficient for large event lists but also never invoked
// in the current system
func (t *events) Delete(oid schema.OID) {
	for _, v := range t.m {
		if v.OID == oid {
			v.deleted = true
		}
	}
}

func (t *events) List() []schema.OID {
	list := []schema.OID{}

	for _, v := range t.m {
		if !v.deleted {
			list = append(list, v.OID)
		}
	}

	return list
}

// FIXME horrifically inefficient for large event lists but also never invoked
//       in the current system
func (t *events) Has(v any, oid schema.OID) bool {
	for _, v := range t.m {
		if v.OID == oid && !v.deleted {
			return true
		}
	}

	return false
}
