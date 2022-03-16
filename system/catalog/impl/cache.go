package memdb

import (
	"fmt"
	"strings"
	"sync"

	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
)

type value interface{}

var cache = struct {
	cache map[schema.OID]value
	guard sync.RWMutex
}{
	cache: map[schema.OID]value{},
}

func Get(oid schema.OID) interface{} {
	cache.guard.RLock()
	defer cache.guard.RUnlock()

	if v, ok := cache.cache[oid]; ok {
		return v
	}

	return nil
}

func (cc *catalog) GetV(oid schema.OID, suffix schema.Suffix) interface{} {
	cache.guard.RLock()
	defer cache.guard.RUnlock()

	if v, ok := cache.cache[oid.Append(suffix)]; ok {
		return v
	}

	return nil
}

func (cc *catalog) Put(oid schema.OID, v interface{}) {
	cache.guard.Lock()
	defer cache.guard.Unlock()

	cache.cache[oid] = v
}

func (cc *catalog) PutV(oid schema.OID, suffix schema.Suffix, v interface{}) {
	cache.guard.Lock()
	defer cache.guard.Unlock()

	cache.cache[oid.Append(suffix)] = v
}

func PutL(objects []schema.Object) {
	if objects != nil && len(objects) > 0 {
		cache.guard.Lock()
		defer cache.guard.Unlock()

		for _, o := range objects {
			cache.cache[o.OID] = o.Value
		}
	}
}

func (cc *catalog) Find(prefix schema.OID, suffix schema.Suffix, value interface{}) (schema.OID, bool) {
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

	return schema.OID(""), false
}

func (cc *catalog) GetDoorDeviceID(door schema.OID) uint32 {
	fields := map[uint8]schema.Suffix{
		1: schema.ControllerDoor1,
		2: schema.ControllerDoor2,
		3: schema.ControllerDoor3,
		4: schema.ControllerDoor4,
	}

	for k, controller := range cc.controllers.m {
		if !controller.deleted {
			for _, s := range fields {
				if v := cc.GetV(k, s); v == door {
					return controller.ID
				}
			}
		}
	}

	return 0
}

func (cc *catalog) GetDoorDeviceDoor(door schema.OID) uint8 {
	fields := map[uint8]schema.Suffix{
		1: schema.ControllerDoor1,
		2: schema.ControllerDoor2,
		3: schema.ControllerDoor3,
		4: schema.ControllerDoor4,
	}

	for k, controller := range cc.controllers.m {
		if !controller.deleted {
			for d, s := range fields {
				if v := cc.GetV(k, s); v == door {
					return d
				}
			}
		}
	}

	return 0
}
