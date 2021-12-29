package events

import (
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
)

const EventDeviceID = catalog.EventDeviceID
const EventIndex = catalog.EventIndex
const EventTimestamp = catalog.EventTimestamp
const EventType = catalog.EventType
const EventDoor = catalog.EventDoor
const EventDirection = catalog.EventDirection
const EventCard = catalog.EventCard
const EventGranted = catalog.EventGranted
const EventReason = catalog.EventReason
const EventDeviceName = catalog.EventDeviceName
const EventDoorName = catalog.EventDoorName
const EventCardName = catalog.EventCardName

var lookup = map[catalog.Suffix]string{
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
