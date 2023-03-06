package events

import (
	"encoding/json"
	"fmt"

	core "github.com/uhppoted/uhppote-core/types"

	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	"github.com/uhppoted/uhppoted-httpd/types"
	"github.com/uhppoted/uhppoted-lib/uhppoted"
)

type Event struct {
	catalog.CatalogEvent
	Timestamp  core.DateTime
	Type       eventType
	Door       uint8
	Direction  direction
	Card       uint32
	Granted    bool
	Reason     reason
	DeviceName string
	DoorName   string
	CardName   string
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

func NewEvent(oid schema.OID, e uhppoted.Event, device, door, card string) Event {
	return Event{
		CatalogEvent: catalog.CatalogEvent{
			OID:      oid,
			DeviceID: e.DeviceID,
			Index:    e.Index,
		},
		Timestamp:  e.Timestamp,
		Type:       eventType(e.Type),
		Door:       e.Door,
		Direction:  direction(e.Direction),
		Card:       e.CardNumber,
		Granted:    e.Granted,
		Reason:     reason(e.Reason),
		DeviceName: device,
		DoorName:   door,
		CardName:   card,
	}
}

func (e Event) IsValid() bool {
	return true
}

func (e Event) IsDeleted() bool {
	return false
}

func (e Event) AsObjects(a auth.OpAuth) []schema.Object {
	type E = struct {
		field schema.Suffix
		value interface{}
	}

	list := []E{}

	list = append(list, E{EventDeviceID, e.DeviceID})
	list = append(list, E{EventIndex, e.Index})
	list = append(list, E{EventTimestamp, e.Timestamp})
	list = append(list, E{EventType, e.Type})
	list = append(list, E{EventDoor, e.Door})
	list = append(list, E{EventDirection, e.Direction})
	list = append(list, E{EventCard, e.Card})
	list = append(list, E{EventReason, e.Reason})
	list = append(list, E{EventGranted, e.Granted})
	list = append(list, E{EventDeviceName, e.DeviceName})
	list = append(list, E{EventDoorName, e.DoorName})
	list = append(list, E{EventCardName, e.CardName})

	objects := []schema.Object{}

	if err := CanView(a, e, "OID", e.OID); err == nil {
		catalog.Join(&objects, catalog.NewObject(e.OID, types.StatusOk))
	}

	for _, v := range list {
		field := lookup[v.field]
		if err := CanView(a, e, field, v.value); err == nil {
			catalog.Join(&objects, catalog.NewObject2(e.OID, v.field, v.value))
		}
	}

	return objects
}

func (e Event) AsRuleEntity() (string, interface{}) {
	entity := struct {
		DeviceID uint32
		Index    uint32
	}{
		DeviceID: e.DeviceID,
		Index:    e.Index,
	}

	return "event", &entity
}

func (e *Event) set(auth auth.OpAuth, oid schema.OID, value string) ([]interface{}, error) {
	objects := []interface{}{}

	return objects, nil
}

func (e Event) serialize() ([]byte, error) {
	record := struct {
		OID        schema.OID    `json:"OID"`
		DeviceID   uint32        `json:"device-id,omitempty"`
		Index      uint32        `json:"index"`
		Timestamp  core.DateTime `json:"timestamp"`
		Type       eventType     `json:"event-type"`
		Door       uint8         `json:"door"`
		Direction  direction     `json:"direction"`
		Card       uint32        `json:"card"`
		Granted    bool          `json:"granted"`
		Reason     reason        `json:"reason"`
		DeviceName string        `json:"device-name,omitempty"`
		DoorName   string        `json:"door-name,omitempty"`
		CardName   string        `json:"card-name,omitempty"`
	}{
		OID:        e.OID,
		DeviceID:   e.DeviceID,
		Index:      e.Index,
		Timestamp:  e.Timestamp,
		Type:       e.Type,
		Door:       e.Door,
		Direction:  e.Direction,
		Card:       e.Card,
		Granted:    e.Granted,
		Reason:     e.Reason,
		DeviceName: e.DeviceName,
		DoorName:   e.DoorName,
		CardName:   e.CardName,
	}

	return json.Marshal(record)
}

func (e *Event) deserialize(bytes []byte) error {
	record := struct {
		OID        schema.OID    `json:"OID"`
		DeviceID   uint32        `json:"device-id"`
		Index      uint32        `json:"index"`
		Timestamp  core.DateTime `json:"timestamp"`
		Type       eventType     `json:"event-type"`
		Door       uint8         `json:"door"`
		Direction  direction     `json:"direction"`
		Card       uint32        `json:"card"`
		Granted    bool          `json:"granted"`
		Reason     reason        `json:"reason"`
		DeviceName string        `json:"device-name"`
		DoorName   string        `json:"door-name"`
		CardName   string        `json:"card-name"`
	}{}

	if err := json.Unmarshal(bytes, &record); err != nil {
		return err
	}

	e.OID = record.OID
	e.DeviceID = record.DeviceID
	e.Index = record.Index
	e.Timestamp = record.Timestamp
	e.Type = record.Type
	e.Door = record.Door
	e.Direction = record.Direction
	e.Card = record.Card
	e.Granted = record.Granted
	e.Reason = record.Reason
	e.DeviceName = record.DeviceName
	e.DoorName = record.DoorName
	e.CardName = record.CardName

	return nil
}

func (e Event) clone() Event {
	event := Event{
		CatalogEvent: catalog.CatalogEvent{
			OID: e.OID,
		},
		Timestamp: e.Timestamp,
	}

	return event
}
