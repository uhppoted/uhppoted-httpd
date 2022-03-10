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

// Poor man's generics :-(
func TypeOf(v interface{}) Type {
	t := fmt.Sprintf("%T", v)
	switch t {
	case "*interfaces.LAN", "interfaces.LAN":
		return TInterface

	case "uint32":
		return TController

	case "*cards.Card", "cards.Card":
		return TCard

	case "*doors.Door", "doors.Door":
		return TDoor

	case "*groups.Group", "groups.Group":
		return TGroup

	case "*events.Event", "events.Event":
		return TEvent

	case "*logs.LogEntry", "logs.LogEntry":
		return TLog

	case "*users.User", "user.User":
		return TUser

	default:
		return TUnknown
	}
}
