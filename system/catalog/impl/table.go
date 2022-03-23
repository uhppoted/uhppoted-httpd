package memdb

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/types"
)

type Table interface {
	New(interface{}) schema.OID
	Put(schema.OID, interface{})
	Delete(schema.OID)
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

func (t *table) Clear() {
	t.m = map[schema.OID]*record{}
	t.last = 0
}

func (t *controllers) New(v interface{}) schema.OID {
	u := v.(ctypes.CatalogController)
	suffix := t.last

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
		ID: v.(uint32),
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

func (t *controllers) Clear() {
	t.m = map[schema.OID]*controller{}
	t.last = 0
}
