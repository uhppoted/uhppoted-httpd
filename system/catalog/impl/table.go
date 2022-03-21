package memdb

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
)

type table[V record] struct {
	base schema.OID
	m    map[schema.OID]V
	last uint32
}

type record interface {
	*entry | *controller
	Delete()
}

func (e *entry) Delete() {
	e.deleted = true
}

type controllers struct {
	base schema.OID
	m    map[schema.OID]*controller
	last uint32
}

type entry struct {
	deleted bool
}

type controller struct {
	deleted bool
	ID      uint32
}

func (c *controller) Delete() {
	c.deleted = true
}

func newOID(t table[*entry], v interface{}) schema.OID {
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

		t.m[oid] = &entry{}
		t.last = suffix
		return oid
	}
}

func put(t table[*entry], oid schema.OID, v interface{}) {
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

	t.m[oid] = &entry{}

	if v := uint32(index); v > t.last {
		t.last = v
	}
}

func (t *table[T]) Delete(oid schema.OID) {
	if v, ok := t.m[oid]; ok {
		v.Delete()
	}
}

func (t *table[pentry]) Clear() {
	t.m = map[schema.OID]pentry{}
	t.last = 0
}

func (t *controllers) New(v uint32) schema.OID {
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

		t.m[oid] = &controller{
			ID: v,
		}
		t.last = suffix
		return oid
	}
}

func (t *controllers) Put(oid schema.OID, v uint32) {
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
		ID: v,
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
