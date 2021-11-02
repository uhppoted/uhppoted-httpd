package catalog

import (
	"fmt"
	"strings"
)

type value interface{}

var cache = map[OID]value{}

func GetV(oid OID) interface{} {
	if v, ok := cache[oid]; ok {
		return v
	}

	return nil
}

func PutV(oid OID, v interface{}) {
	cache[oid] = v
}

func Find(prefix OID, suffix Suffix, value interface{}) (OID, bool) {
	s := fmt.Sprintf("%v", value)

	for k, v := range cache {
		prefixed := strings.HasPrefix(string(k), string(prefix))
		suffixed := strings.HasSuffix(string(k), string(suffix))
		if prefixed && suffixed && s == fmt.Sprintf("%v", v) {
			return k, true
		}
	}

	return OID(""), false
}
