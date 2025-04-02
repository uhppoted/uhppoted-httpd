package interfaces

import (
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
)

const LANStatus = schema.Status
const LANCreated = schema.Created
const LANDeleted = schema.Deleted
const LANType = schema.Type

const LANName = schema.InterfaceName
const LANBindAddress = schema.LANBindAddress
const LANBroadcastAddress = schema.LANBroadcastAddress
const LANListenAddress = schema.LANListenAddress

const ControllerTouched = schema.Touched
const ControllerEndpointAddress = schema.ControllerEndpointAddress
const ControllerDateTimeCurrent = schema.ControllerDateTimeCurrent
const ControllerDateTimeModified = schema.ControllerDateTimeModified
const ControllerCardsCount = schema.ControllerCardsCount
const ControllerCardsStatus = schema.ControllerCardsStatus
const ControllerEventsStatus = schema.ControllerEventsStatus
const ControllerEventsFirst = schema.ControllerEventsFirst
const ControllerEventsLast = schema.ControllerEventsLast
const ControllerEventsCurrent = schema.ControllerEventsCurrent
const ControllerAntiPassback = schema.ControllerAntiPassback

const DoorDelay = schema.DoorDelay
const DoorDelayConfigured = schema.DoorDelayConfigured
const DoorDelayModified = schema.DoorDelayModified
const DoorControl = schema.DoorControl
const DoorControlModified = schema.DoorControlModified
const DoorControlConfigured = schema.DoorControlConfigured

var lookup = map[schema.Suffix]string{
	LANStatus:           "LAN.status",
	LANCreated:          "LAN.created",
	LANDeleted:          "LAN.deleted",
	LANType:             "LAN.type",
	LANName:             "LAN.name",
	LANBindAddress:      "LAN.address.bind",
	LANBroadcastAddress: "LAN.address.broadcast",
	LANListenAddress:    "LAN.address.listen",
}
