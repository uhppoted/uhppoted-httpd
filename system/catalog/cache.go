package catalog

import ()

type value struct {
	value interface{}
	dirty bool
}

var cache = map[string]value{}

func GetV(oid string) (interface{}, bool) {
	v, ok := cache[oid]
	if ok {
		return v.value, v.dirty
	}

	return nil, false
}

func PutV(oid string, v interface{}, dirty bool) {
	cache[oid] = value{
		value: v,
		dirty: dirty,
	}
}
