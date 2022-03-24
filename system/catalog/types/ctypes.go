package ctypes

import (
	"fmt"

	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
)

type CatalogType interface {
	CatalogInterface | CatalogController | CatalogDoor | CatalogCard | CatalogGroup | CatalogEvent | CatalogLogEntry | CatalogUser
}

type CatalogInterface struct {
}

type CatalogController struct {
	OID      schema.OID
	DeviceID uint32
}

type CatalogDoor struct {
}

type CatalogCard struct {
}

type CatalogGroup struct {
}

type CatalogEvent struct {
}

type CatalogLogEntry struct {
}

type CatalogUser struct {
}

type Type int

const (
	TUnknown Type = iota
	TInterface
	TController
	TDoor
	TCard
	TGroup
	TEvent
	TLog
	TUser
)

func (t Type) String() string {
	return []string{
		"unknown",
		"interface",
		"controller",
		"door",
		"card",
		"group",
		"event",
		"log",
		"user",
	}[t]
}

// ... pending real Go generics
func TypeOf(v interface{}) Type {
	t := fmt.Sprintf("%T", v)
	switch t {
	case "ctypes.CatalogInterface":
		return TInterface

	case "ctypes.CatalogController":
		return TController

	case "ctypes.CatalogCard":
		return TCard

	case "ctypes.CatalogDoor":
		return TDoor

	case "ctypes.CatalogGroup":
		return TGroup

	case "ctypes.CatalogEvent":
		return TEvent

	case "ctypes.CatalogLogEntry":
		return TLog

	case "ctypes.CatalogUser":
		return TUser

	default:
		return TUnknown
	}
}
