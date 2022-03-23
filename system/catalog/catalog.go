package catalog

import (
	"fmt"

	"github.com/uhppoted/uhppoted-httpd/system/catalog/impl"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/types"
)

type Catalog interface {
	Clear()

	NewT(ctypes.Type, interface{}) schema.OID
	PutT(ctypes.Type, interface{}, schema.OID)
	DeleteT(ctypes.Type, schema.OID)
	ListT(ctypes.Type) []schema.OID
	HasT(ctypes.Type, schema.OID) bool

	GetV(schema.OID, schema.Suffix) interface{}
	Put(schema.OID, interface{})
	PutV(schema.OID, schema.Suffix, interface{})

	Find(prefix schema.OID, suffix schema.Suffix, value interface{}) (schema.OID, bool)
	FindController(deviceID uint32) schema.OID

	GetDoorDeviceID(door schema.OID) uint32
	GetDoorDeviceDoor(door schema.OID) uint8
}

var catalog Catalog = memdb.Catalog()

func Join(p *[]schema.Object, q ...schema.Object) {
	*p = append(*p, q...)
}

func NewObject(oid schema.OID, value interface{}) schema.Object {
	return schema.Object{
		OID:   oid,
		Value: value,
	}
}

func NewObject2(oid schema.OID, suffix schema.Suffix, value interface{}) schema.Object {
	return schema.Object{
		OID:   oid.Append(suffix),
		Value: value,
	}
}

func Clear() {
	catalog.Clear()
}

func NewT[T ctypes.CatalogType](v T) schema.OID {
	if t := ctypes.TypeOf(v); t == ctypes.TUnknown {
		panic(fmt.Sprintf("Unsupported catalog type: %T", v))
	} else {
		return catalog.NewT(t, v)
	}
}

func PutT(v interface{}, oid schema.OID) {
	if t := ctypes.TypeOf(v); t == ctypes.TUnknown {
		panic(fmt.Sprintf("Unsupported catalog type: %T", v))
	} else {
		catalog.PutT(t, v, oid)
	}
}

func DeleteT(v interface{}, oid schema.OID) {
	if t := ctypes.TypeOf(v); t == ctypes.TUnknown {
		panic(fmt.Sprintf("Unsupported catalog type: %T", v))
	} else {
		catalog.DeleteT(t, oid)
	}
}

func GetV(oid schema.OID, suffix schema.Suffix) interface{} {
	return catalog.GetV(oid, suffix)
}

func GetBool(oid schema.OID, suffix schema.Suffix) (bool, bool) {
	if v := catalog.GetV(oid, suffix); v == nil {
		return false, false
	} else if b, ok := v.(bool); !ok {
		return false, false
	} else {
		return b, true
	}
}
func GetUint8(oid schema.OID, suffix schema.Suffix) (uint8, bool) {
	if v := catalog.GetV(oid, suffix); v == nil {
		return 0, false
	} else if u, ok := v.(uint8); !ok {
		return 0, false
	} else {
		return u, true
	}
}

func Put(oid schema.OID, v interface{}) {
	catalog.Put(oid, v)
}

func PutV(oid schema.OID, suffix schema.Suffix, v interface{}) {
	catalog.PutV(oid, suffix, v)
}

func Find(prefix schema.OID, suffix schema.Suffix, value interface{}) (schema.OID, bool) {
	return catalog.Find(prefix, suffix, value)
}

func FindController(deviceID uint32) schema.OID {
	return catalog.FindController(deviceID)
}

func GetDoors() []schema.OID {
	return catalog.ListT(ctypes.TDoor)
}

func GetDoorDeviceID(door schema.OID) uint32 {
	return catalog.GetDoorDeviceID(door)
}

func GetDoorDeviceDoor(door schema.OID) uint8 {
	return catalog.GetDoorDeviceDoor(door)
}

func GetGroups() []schema.OID {
	return catalog.ListT(ctypes.TGroup)
}

func HasGroup(oid schema.OID) bool {
	return catalog.HasT(ctypes.TGroup, oid)
}
