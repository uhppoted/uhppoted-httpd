package events

import (
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
)

const EventDeviceID = schema.EventDeviceID
const EventIndex = schema.EventIndex
const EventTimestamp = schema.EventTimestamp
const EventType = schema.EventType
const EventDoor = schema.EventDoor
const EventDirection = schema.EventDirection
const EventCard = schema.EventCard
const EventGranted = schema.EventGranted
const EventReason = schema.EventReason
const EventDeviceName = schema.EventDeviceName
const EventDoorName = schema.EventDoorName
const EventCardName = schema.EventCardName

var lookup = map[schema.Suffix]string{
	EventDeviceID:   "event.device.ID",
	EventIndex:      "event.index",
	EventTimestamp:  "event.timestamp",
	EventType:       "event.type",
	EventDoor:       "event.door",
	EventDirection:  "event.direction",
	EventCard:       "event.card",
	EventGranted:    "event.granted",
	EventReason:     "event.reason",
	EventDeviceName: "event.device.name",
	EventDoorName:   "event.door.name",
	EventCardName:   "event.card.name",
}
