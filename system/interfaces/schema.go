package interfaces

import (
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
)

const LANStatus = catalog.Status
const LANCreated = catalog.Created
const LANDeleted = catalog.Deleted
const LANType = catalog.Type

const LANName = catalog.InterfaceName
const LANBindAddress = catalog.LANBindAddress
const LANBroadcastAddress = catalog.LANBroadcastAddress
const LANListenAddress = catalog.LANListenAddress

const ControllerTouched = catalog.Touched
const ControllerEndpointAddress = catalog.ControllerEndpointAddress
const ControllerDateTimeCurrent = catalog.ControllerDateTimeCurrent
const ControllerCardsCount = catalog.ControllerCardsCount
const ControllerCardsStatus = catalog.ControllerCardsStatus
const ControllerEventsStatus = catalog.ControllerEventsStatus
const ControllerEventsFirst = catalog.ControllerEventsFirst
const ControllerEventsLast = catalog.ControllerEventsLast
const ControllerEventsCurrent = catalog.ControllerEventsCurrent

const DoorDelay = catalog.DoorDelay
const DoorDelayModified = catalog.DoorDelayModified
const DoorControl = catalog.DoorControl
const DoorControlModified = catalog.DoorControlModified

var lookup = map[catalog.Suffix]string{
	LANStatus:           "LAN.status",
	LANCreated:          "LAN.created",
	LANDeleted:          "LAN.deleted",
	LANType:             "LAN.type",
	LANName:             "LAN.name",
	LANBindAddress:      "LAN.address.bind",
	LANBroadcastAddress: "LAN.address.broadcast",
	LANListenAddress:    "LAN.address.listen",
}
