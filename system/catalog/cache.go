package catalog

import (
	"fmt"
	"strings"
)

type value struct {
	value interface{}
	dirty bool
}

var cache = map[OID]value{}

func GetV(oid OID) (interface{}, bool) {
	v, ok := cache[oid]
	if ok {
		return v.value, v.dirty
	}

	return nil, false
}

func PutV(oid OID, v interface{}, dirty bool) {
	cache[oid] = value{
		value: v,
		dirty: dirty,
	}
}

func Find(prefix OID, suffix Suffix, value interface{}) (OID, bool) {
	s := fmt.Sprintf("%v", value)

	for k, v := range cache {
		prefixed := strings.HasPrefix(string(k), string(prefix))
		suffixed := strings.HasSuffix(string(k), string(suffix))
		if prefixed && suffixed && s == fmt.Sprintf("%v", v.value) {
			return k, true
		}
	}

	return OID(""), false
}
