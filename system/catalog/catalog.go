package catalog

import (
	"fmt"

	"github.com/uhppoted/uhppoted-httpd/system/catalog/impl"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/types"
)

type Catalog interface {
	Clear()
	Delete(oid schema.OID)

	GetV(oid schema.OID, suffix schema.Suffix) interface{}
	Put(oid schema.OID, v interface{})
	PutV(oid schema.OID, suffix schema.Suffix, v interface{})
	Find(prefix schema.OID, suffix schema.Suffix, value interface{}) (schema.OID, bool)

	NewController(deviceID uint32) schema.OID
	NewCard() schema.OID
	NewDoor() schema.OID
	NewGroup() schema.OID
	NewEvent() schema.OID
	NewLogEntry() schema.OID
	NewUser() schema.OID

	PutT(t ctypes.Type, v interface{}, oid schema.OID)

	FindController(deviceID uint32) schema.OID

	Doors() map[schema.OID]struct{}
	GetDoorDeviceID(door schema.OID) uint32
	GetDoorDeviceDoor(door schema.OID) uint8

	Groups() map[schema.OID]struct{}
	HasGroup(oid schema.OID) bool
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

func Delete(oid schema.OID) {
	catalog.Delete(oid)
}

func GetV(oid schema.OID, suffix schema.Suffix) interface{} {
	return catalog.GetV(oid, suffix)
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

func NewController(deviceID uint32) schema.OID {
	return catalog.NewController(deviceID)
}

func NewCard() schema.OID {
	return catalog.NewCard()
}

func NewDoor() schema.OID {
	return catalog.NewDoor()
}

func NewGroup() schema.OID {
	return catalog.NewGroup()
}

func NewEvent() schema.OID {
	return catalog.NewEvent()
}

func NewLogEntry() schema.OID {
	return catalog.NewLogEntry()
}

func NewUser() schema.OID {
	return catalog.NewUser()
}

func PutT(v interface{}, oid schema.OID) {
	if t := ctypes.TypeOf(v); t != ctypes.TUnknown {
		catalog.PutT(t, v, oid)
		return
	}

	panic(fmt.Sprintf("Unknown catalog type: %T", v))
}

func FindController(deviceID uint32) schema.OID {
	return catalog.FindController(deviceID)
}

func GetDoors() []schema.OID {
	list := []schema.OID{}
	doors := catalog.Doors()

	for d, _ := range doors {
		list = append(list, d)
	}

	return list
}

func GetDoorDeviceID(door schema.OID) uint32 {
	return catalog.GetDoorDeviceID(door)
}

func GetDoorDeviceDoor(door schema.OID) uint8 {
	return catalog.GetDoorDeviceDoor(door)
}

func GetGroups() []schema.OID {
	list := []schema.OID{}
	groups := catalog.Groups()

	for g, _ := range groups {
		list = append(list, g)
	}

	return list
}

func HasGroup(oid schema.OID) bool {
	return catalog.HasGroup(oid)
}
