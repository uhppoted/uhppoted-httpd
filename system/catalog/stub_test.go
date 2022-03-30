package catalog

import (
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
)

type stub struct {
}

func (s stub) NewT(any) schema.OID {
	return ""
}

func (s stub) PutT(any, schema.OID) {
}

func (s stub) DeleteT(any, schema.OID) {
}

func (s stub) ListT(schema.OID) []schema.OID {
	return nil
}

func (s stub) HasT(any, schema.OID) bool {
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
