package catalog

import ()

type value struct {
	value interface{}
	dirty bool
}

var cache = map[string]value{}

func GetV(oid OID) (interface{}, bool) {
	v, ok := cache[string(oid)]
	if ok {
		return v.value, v.dirty
	}

	return nil, false
}

func PutV(oid OID, v interface{}, dirty bool) {
	cache[string(oid)] = value{
		value: v,
		dirty: dirty,
	}
}
