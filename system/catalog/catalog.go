package catalog

import (
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
)

type Catalog interface {
	NewT(any) schema.OID
	PutT(any, schema.OID)
	DeleteT(any, schema.OID)
	ListT(schema.OID) []schema.OID
	HasT(any, schema.OID) bool

	GetV(schema.OID, schema.Suffix) any
	Put(schema.OID, any)
	PutV(schema.OID, schema.Suffix, any)

	Find(prefix schema.OID, suffix schema.Suffix, value any) (schema.OID, bool)
	FindController(v CatalogController) schema.OID

	GetDoorDeviceID(door schema.OID) uint32
	GetDoorDeviceDoor(door schema.OID) uint8
}

type CatalogType interface {
	CatalogInterface |
		CatalogController |
		CatalogDoor |
		CatalogCard |
		CatalogGroup |
		CatalogEvent |
		CatalogLogEntry |
		CatalogUser

	oid() schema.OID
}

var catalog Catalog

func Init(c Catalog) {
	catalog = c
}

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

func NewT[T CatalogType](v T) schema.OID {
	return catalog.NewT(v)
}

func PutT[T CatalogType](v T) {
	oid := v.oid()

	catalog.PutT(v, oid)
}

func DeleteT[T CatalogType](v T, oid schema.OID) {
	catalog.DeleteT(v, oid)
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

func Put(oid schema.OID, v any) {
	catalog.Put(oid, v)
}

func PutV(oid schema.OID, suffix schema.Suffix, v interface{}) {
	catalog.PutV(oid, suffix, v)
}

func Find(prefix schema.OID, suffix schema.Suffix, value interface{}) (schema.OID, bool) {
	return catalog.Find(prefix, suffix, value)
}

func FindController(deviceID uint32) schema.OID {
	return catalog.FindController(CatalogController{DeviceID: deviceID})
}

func GetDoors() []schema.OID {
	return catalog.ListT(schema.DoorsOID)
}

func GetDoorDeviceID(door schema.OID) uint32 {
	return catalog.GetDoorDeviceID(door)
}

func GetDoorDeviceDoor(door schema.OID) uint8 {
	return catalog.GetDoorDeviceDoor(door)
}

func GetGroups() []schema.OID {
	return catalog.ListT(schema.GroupsOID)
}

func HasGroup(oid schema.OID) bool {
	type group struct {
		CatalogGroup
	}

	return catalog.HasT(group{}.CatalogGroup, oid)
}
