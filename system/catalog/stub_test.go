package catalog

import (
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/types"
)

type stub struct {
}

func (s stub) Clear() {
}

func (s stub) NewT(ctypes.Type, interface{}) schema.OID {
	return ""
}

func (s stub) PutT(ctypes.Type, interface{}, schema.OID) {
}

func (s stub) DeleteT(ctypes.Type, schema.OID) {
}

func (s stub) ListT(ctypes.Type) []schema.OID {
	return nil
}

func (s stub) HasT(ctypes.Type, schema.OID) bool {
	return false
}

func (s stub) GetV(schema.OID, schema.Suffix) interface{} {
	return nil
}

func (s stub) Put(schema.OID, interface{}) {
}

func (s stub) PutV(schema.OID, schema.Suffix, interface{}) {
}

func (s stub) Find(prefix schema.OID, suffix schema.Suffix, value interface{}) (schema.OID, bool) {
	return "", false
}

func (s stub) FindController(deviceID uint32) schema.OID {
	return ""
}

func (s stub) GetDoorDeviceID(door schema.OID) uint32 {
	return 0
}

func (s stub) GetDoorDeviceDoor(door schema.OID) uint8 {
	return 0
}
