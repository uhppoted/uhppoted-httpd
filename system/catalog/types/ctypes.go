package ctypes

import (
	"fmt"
)

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

	case "*controllers.Controller", "uint32", "catalog.CatalogController":
		return TController

	case "*cards.Card", "cards.Card", "catalog.CatalogCard":
		return TCard

	case "*doors.Door", "doors.Door", "catalog.CatalogDoor":
		return TDoor

	case "*groups.Group", "groups.Group", "catalog.CatalogGroup":
		return TGroup

	case "*events.Event", "events.Event", "catalog.CatalogEvent":
		return TEvent

	case "*logs.LogEntry", "logs.LogEntry", "catalog.CatalogLogEntry":
		return TLog

	case "*users.User", "user.User", "catalog.CatalogUser":
		return TUser

	default:
		return TUnknown
	}
}
