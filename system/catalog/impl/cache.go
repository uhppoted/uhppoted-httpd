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
	sync.RWMutex
}{
	cache: map[schema.OID]value{},
}

func Get(oid schema.OID) interface{} {
	cache.RLock()
	defer cache.RUnlock()

	if v, ok := cache.cache[oid]; ok {
		return v
	}

	return nil
}

func (cc *db) GetV(oid schema.OID, suffix schema.Suffix) interface{} {
	cache.RLock()
	defer cache.RUnlock()

	if v, ok := cache.cache[oid.Append(suffix)]; ok {
		return v
	}

	return nil
}

func (cc *db) Put(oid schema.OID, v interface{}) {
	cache.Lock()
	defer cache.Unlock()

	cache.cache[oid] = v
}

func (cc *db) PutV(oid schema.OID, suffix schema.Suffix, v interface{}) {
	cache.Lock()
	defer cache.Unlock()

	cache.cache[oid.Append(suffix)] = v
}

func PutL(objects []schema.Object) {
	if len(objects) > 0 {
		cache.Lock()
		defer cache.Unlock()

		for _, o := range objects {
			cache.cache[o.OID] = o.Value
		}
	}
}

func (cc *db) Find(prefix schema.OID, suffix schema.Suffix, value interface{}) (schema.OID, bool) {
	cache.RLock()
	defer cache.RUnlock()

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

func (cc *db) GetDoorDeviceID(door schema.OID) uint32 {
	fields := map[uint8]schema.Suffix{
		1: schema.ControllerDoor1,
		2: schema.ControllerDoor2,
		3: schema.ControllerDoor3,
		4: schema.ControllerDoor4,
	}

	for k, controller := range cc.controllers.(*controllers).m {
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

func (cc *db) GetDoorDeviceDoor(door schema.OID) uint8 {
	fields := map[uint8]schema.Suffix{
		1: schema.ControllerDoor1,
		2: schema.ControllerDoor2,
		3: schema.ControllerDoor3,
		4: schema.ControllerDoor4,
	}

	for k, controller := range cc.controllers.(*controllers).m {
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
