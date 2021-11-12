package catalog

import (
	"fmt"
	"strings"
	"sync"
)

type value interface{}

var cache = struct {
	cache map[OID]value
	guard sync.RWMutex
}{
	cache: map[OID]value{},
}

func Get(oid OID) interface{} {
	cache.guard.RLock()
	defer cache.guard.RUnlock()

	if v, ok := cache.cache[oid]; ok {
		return v
	}

	return nil
}

func GetV(oid OID, suffix Suffix) interface{} {
	cache.guard.RLock()
	defer cache.guard.RUnlock()

	if v, ok := cache.cache[oid.Append(suffix)]; ok {
		return v
	}

	return nil
}

func Put(oid OID, v interface{}) {
	cache.guard.Lock()
	defer cache.guard.Unlock()

	cache.cache[oid] = v
}

func PutV(oid OID, suffix Suffix, v interface{}) {
	cache.guard.Lock()
	defer cache.guard.Unlock()

	cache.cache[oid.Append(suffix)] = v
}

func PutL(objects []Object) {
	if objects != nil && len(objects) > 0 {
		cache.guard.Lock()
		defer cache.guard.Unlock()

		for _, o := range objects {
			cache.cache[o.OID] = o.Value
		}
	}
}

func Find(prefix OID, suffix Suffix, value interface{}) (OID, bool) {
	cache.guard.RLock()
	defer cache.guard.RUnlock()

	s := fmt.Sprintf("%v", value)

	for k, v := range cache.cache {
		prefixed := strings.HasPrefix(string(k), string(prefix))
		suffixed := strings.HasSuffix(string(k), string(suffix))
		if prefixed && suffixed && s == fmt.Sprintf("%v", v) {
			return k, true
		}
	}

	return OID(""), false
}
