package memdb

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
)

type Table interface {
	New(interface{}) schema.OID
	Put(schema.OID, interface{})
	Delete(schema.OID)
	List() []schema.OID
	Has(v any, oid schema.OID) bool
	Clear()
}

type table struct {
	base schema.OID
	m    map[schema.OID]*record
	last uint32
}

type controllers struct {
	base schema.OID
	m    map[schema.OID]*controller
	last uint32
}

type record struct {
	deleted bool
}

type controller struct {
	deleted bool
	ID      uint32
}

func (t *table) New(v interface{}) schema.OID {
	suffix := t.last

loop:
	for {
		suffix += 1
		oid := schema.OID(fmt.Sprintf("%v.%d", t.base, suffix))
		for v, _ := range t.m {
			if v == oid {
				continue loop
			}
		}

		t.m[oid] = &record{}
		t.last = suffix
		return oid
	}
}

func (t *table) Put(oid schema.OID, v interface{}) {
	if !oid.HasPrefix(t.base) {
		panic(fmt.Sprintf("PUT: illegal oid %v for base %v", oid, t.base))
	}

	suffix := strings.TrimPrefix(string(oid), string(t.base))

	match := regexp.MustCompile(`\.([0-9]+)`).FindStringSubmatch(suffix)
	if match == nil || len(match) != 2 {
		panic(fmt.Sprintf("PUT: invalid oid %v for base %v", oid, t.base))
	}

	index, err := strconv.ParseUint(match[1], 10, 32)
	if err != nil {
		panic(fmt.Sprintf("PUT: out of range oid %v for base %v", oid, t.base))
	}

	t.m[oid] = &record{}

	if v := uint32(index); v > t.last {
		t.last = v
	}
}

func (t *table) Delete(oid schema.OID) {
	if v, ok := t.m[oid]; ok {
		v.deleted = true
	}
}

func (t *table) List() []schema.OID {
	list := []schema.OID{}

	for d, v := range t.m {
		if !v.deleted {
			list = append(list, d)
		}
	}

	return list
}

func (t *table) Has(v any, oid schema.OID) bool {
	if v, ok := t.m[oid]; ok && !v.deleted {
		return true
	}

	return false
}

func (t *table) Clear() {
	t.m = map[schema.OID]*record{}
	t.last = 0
}

func (t *controllers) New(v interface{}) schema.OID {
	u := v.(catalog.CatalogController)
	suffix := t.last

	if deviceID := u.DeviceID; deviceID != 0 {
		for oid, c := range t.m {
			if !c.deleted && c.ID == deviceID {
				return oid
			}
		}
	}

loop:
	for {
		suffix += 1
		oid := schema.OID(fmt.Sprintf("%v.%d", t.base, suffix))
		for c, _ := range t.m {
			if c == oid {
				continue loop
			}
		}

		t.m[oid] = &controller{
			ID: u.DeviceID,
		}
		t.last = suffix
		return oid
	}
}

func (t *controllers) Put(oid schema.OID, v interface{}) {
	if !oid.HasPrefix(t.base) {
		panic(fmt.Sprintf("PUT: illegal oid %v for base %v", oid, t.base))
	}

	suffix := strings.TrimPrefix(string(oid), string(t.base))

	match := regexp.MustCompile(`\.([0-9]+)`).FindStringSubmatch(suffix)
	if match == nil || len(match) != 2 {
		panic(fmt.Sprintf("PUT: invalid oid %v for base %v", oid, t.base))
	}

	index, err := strconv.ParseUint(match[1], 10, 32)
	if err != nil {
		panic(fmt.Sprintf("PUT: out of range oid %v for base %v", oid, t.base))
	}

	t.m[oid] = &controller{
		ID: v.(catalog.CatalogController).DeviceID,
	}

	if v := uint32(index); v > t.last {
		t.last = v
	}
}

func (t *controllers) Delete(oid schema.OID) {
	if v, ok := t.m[oid]; ok {
		v.deleted = true
		t.m[oid] = v
	}
}

func (t *controllers) List() []schema.OID {
	list := []schema.OID{}

	for d, v := range t.m {
		if !v.deleted {
			list = append(list, d)
		}
	}

	return list
}

func (t *controllers) Has(v any, oid schema.OID) bool {
	if v, ok := t.m[oid]; ok && !v.deleted {
		return true
	}

	return false
}

func (t *controllers) Clear() {
	t.m = map[schema.OID]*controller{}
	t.last = 0
}
