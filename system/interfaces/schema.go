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
const ControllerEventsCount = catalog.ControllerEventsCount

const DoorDelay = catalog.DoorDelay
const DoorControl = catalog.DoorControl
