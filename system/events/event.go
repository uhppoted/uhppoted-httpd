package events

import (
	"fmt"

	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/types"
	"github.com/uhppoted/uhppoted-lib/uhppoted"
)

type Event struct {
	OID       catalog.OID    `json:"OID"`
	DeviceID  uint32         `json:"device-id"`
	Index     uint32         `json:"index"`
	Timestamp types.DateTime `json:"timestamp"`
	Type      eventType      `json:"event-type"`
	Door      uint8          `json:"door"`
	Direction direction      `json:"direction"`
	Card      uint32         `json:"card"`
	Granted   bool           `json:"granted"`
	Reason    reason         `json:"reason"`
	DoorName  string         `json:"door-name"`
	CardName  string         `json:"card-name"`
}

type eventType uint8

func (e eventType) String() string {
	switch e {
	case 1:
		return "swipe"
	case 2:
		return "door"
	case 3:
		return "warning"
	}

	return fmt.Sprintf("%v", uint8(e))
}

type direction uint8

func (d direction) String() string {
	switch d {
	case 1:
		return "in"
	case 2:
		return "out"
	}

	return ""
}

type reason uint8

func (r reason) String() string {
	switch r {
	case 1:
		return "swipe"
	case 5:
		return "host control"
	case 6:
		return "no privilege"
	case 7:
		return "invalid password"
	case 8:
		return "anti-passback"
	case 9:
		return "more cards"
	case 10:
		return "first card open"
	case 11:
		return "normally closed"
	case 12:
		return "interlock"
	case 13:
		return "time profile"
	case 15:
		return "invalid timezone"
	case 18:
		return "access denied"
	case 20:
		return "push button"
	case 23:
		return "door open"
	case 24:
		return "door closed"
	case 25:
		return "supervisor override"
	case 28:
		return "power on"
	case 29:
		return "reset"
	case 31:
		return "push button: forced lock"
	case 32:
		return "push button: not online"
	case 33:
		return "push button: interlock"
	case 34:
		return "threat"
	case 37:
		return "open too long"
	case 38:
		return "forced open"
	case 39:
		return "fire"
	case 40:
		return "forced close"
	case 41:
		return "anti-theft"
	case 42:
		return "24/7 zone"
	case 43:
		return "emergency call"
	case 44:
		return "remote open"
	case 45:
		return "USB reader open"
	}

	return ""
}

const EventDeviceID = catalog.EventDeviceID
const EventIndex = catalog.EventIndex
const EventTimestamp = catalog.EventTimestamp
const EventType = catalog.EventType
const EventDoor = catalog.EventDoor
const EventDirection = catalog.EventDirection
const EventCard = catalog.EventCard
const EventGranted = catalog.EventGranted
const EventReason = catalog.EventReason
const EventDoorName = catalog.EventDoorName
const EventCardName = catalog.EventCardName

func NewEvent(oid catalog.OID, e uhppoted.Event, lookup func(uhppoted.Event) (string, string)) Event {
	door, card := lookup(e)

	return Event{
		OID:       oid,
		DeviceID:  e.DeviceID,
		Index:     e.Index,
		Timestamp: types.DateTime(e.Timestamp),
		Type:      eventType(e.Type),
		Door:      e.Door,
		Direction: direction(e.Direction),
		Card:      e.CardNumber,
		Granted:   e.Granted,
		Reason:    reason(e.Reason),
		DoorName:  door,
		CardName:  card,
	}
}

func (e Event) IsValid() bool {
	return true
}

func (e Event) IsDeleted() bool {
	return false
}

func (e *Event) AsObjects() []interface{} {
	objects := []interface{}{
		catalog.NewObject(e.OID, types.StatusOk),
		catalog.NewObject2(e.OID, EventDeviceID, e.DeviceID),
		catalog.NewObject2(e.OID, EventTimestamp, e.Timestamp),
		catalog.NewObject2(e.OID, EventType, e.Type),
		catalog.NewObject2(e.OID, EventDoor, e.Door),
		catalog.NewObject2(e.OID, EventDirection, e.Direction),
		catalog.NewObject2(e.OID, EventCard, e.Card),
		catalog.NewObject2(e.OID, EventReason, e.Reason),
		catalog.NewObject2(e.OID, EventGranted, e.Granted),
		catalog.NewObject2(e.OID, EventDoorName, e.DoorName),
		catalog.NewObject2(e.OID, EventCardName, e.CardName),
	}

	return objects
}

func (e Event) clone() Event {
	event := Event{
		OID:       e.OID,
		Timestamp: e.Timestamp,
	}

	return event
}

func (e *Event) set(auth auth.OpAuth, oid catalog.OID, value string) ([]interface{}, error) {
	objects := []interface{}{}

	return objects, nil
}
