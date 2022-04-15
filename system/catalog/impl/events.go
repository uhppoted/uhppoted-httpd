package memdb

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
)

type events struct {
	base schema.OID
	m    map[schema.OID]*event
	last uint32
}

type event struct {
	deleted bool
}

func (t *events) New(v any) schema.OID {
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

		t.m[oid] = &event{}
		t.last = suffix
		return oid
	}
}

func (t *events) Put(oid schema.OID, v any) {
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

	t.m[oid] = &event{}

	if v := uint32(index); v > t.last {
		t.last = v
	}
}

func (t *events) Delete(oid schema.OID) {
	if v, ok := t.m[oid]; ok {
		v.deleted = true
	}
}

func (t *events) List() []schema.OID {
	list := []schema.OID{}

	for d, v := range t.m {
		if !v.deleted {
			list = append(list, d)
		}
	}

	return list
}

func (t *events) Has(v any, oid schema.OID) bool {
	if v, ok := t.m[oid]; ok && !v.deleted {
		return true
	}

	return false
}
