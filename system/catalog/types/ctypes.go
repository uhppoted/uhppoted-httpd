package ctypes

import (
	"fmt"
)

type CatalogType interface {
	CatalogInterface | CatalogController | CatalogDoor | CatalogCard | CatalogGroup | CatalogEvent | CatalogLogEntry | CatalogUser
}

type CatalogInterface struct {
}

type CatalogController struct {
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
	case "*interfaces.LAN", "interfaces.LAN", "catalog.CatalogInterface":
		return TInterface

	case "*controllers.Controller", "uint32", "ctypes.CatalogController":
		return TController

	case "*cards.Card", "cards.Card", "ctypes.CatalogCard":
		return TCard

	case "*doors.Door", "doors.Door", "ctypes.CatalogDoor":
		return TDoor

	case "*groups.Group", "groups.Group", "ctypes.CatalogGroup":
		return TGroup

	case "*events.Event", "events.Event", "ctypes.CatalogEvent":
		return TEvent

	case "*logs.LogEntry", "logs.LogEntry", "ctypes.CatalogLogEntry":
		return TLog

	case "*users.User", "user.User", "ctypes.CatalogUser":
		return TUser

	default:
		return TUnknown
	}
}
