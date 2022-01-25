package controllers

import (
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
)

const LANStatus = catalog.Status
const LANType = catalog.Type
const LANName = catalog.InterfaceName
const LANBindAddress = catalog.LANBindAddress
const LANBroadcastAddress = catalog.LANBroadcastAddress
const LANListenAddress = catalog.LANListenAddress
const DoorDelay = catalog.DoorDelay
const DoorDelayModified = catalog.DoorDelayModified
const DoorControl = catalog.DoorControl
const DoorControlModified = catalog.DoorControlModified

const ControllerStatus = catalog.Status
const ControllerCreated = catalog.Created
const ControllerDeleted = catalog.Deleted
const ControllerTouched = catalog.Touched
const ControllerName = catalog.ControllerName
const ControllerDeviceID = catalog.ControllerDeviceID
const ControllerEndpointStatus = catalog.ControllerEndpointStatus
const ControllerEndpointAddress = catalog.ControllerEndpointAddress
const ControllerEndpointConfigured = catalog.ControllerEndpointConfigured
const ControllerDateTime = catalog.ControllerDateTime
const ControllerDateTimeStatus = catalog.ControllerDateTimeStatus
const ControllerDateTimeCurrent = catalog.ControllerDateTimeCurrent
const ControllerDateTimeSystem = catalog.ControllerDateTimeSystem
const ControllerDateTimeModified = catalog.ControllerDateTimeModified
const ControllerCardsStatus = catalog.ControllerCardsStatus
const ControllerCardsCount = catalog.ControllerCardsCount
const ControllerEventsFirst = catalog.ControllerEventsFirst
const ControllerEventsLast = catalog.ControllerEventsLast
const ControllerEventsCurrent = catalog.ControllerEventsCurrent
const ControllerEventsStatus = catalog.ControllerEventsStatus
const ControllerDoor1 = catalog.ControllerDoor1
const ControllerDoor2 = catalog.ControllerDoor2
const ControllerDoor3 = catalog.ControllerDoor3
const ControllerDoor4 = catalog.ControllerDoor4

var lookup = map[catalog.Suffix]string{
	ControllerStatus:             "controller.status",
	ControllerCreated:            "controller.created",
	ControllerDeleted:            "controller.deleted",
	ControllerTouched:            "controller.touched",
	ControllerName:               "controller.name",
	ControllerDeviceID:           "controller.device.ID",
	ControllerEndpointStatus:     "controller.endpoint.status",
	ControllerEndpointAddress:    "controller.endpoint.address",
	ControllerEndpointConfigured: "controller.endpoint.configured",
	ControllerDateTime:           "controller.datetime",
	ControllerDateTimeStatus:     "controller.datetime.status",
	ControllerDateTimeCurrent:    "controller.datetime.current",
	ControllerDateTimeSystem:     "controller.datetime.system",
	ControllerDateTimeModified:   "controller.datetime.modified",
	ControllerCardsStatus:        "controller.cards.status",
	ControllerCardsCount:         "controller.cards.count",
	ControllerEventsFirst:        "controller.events.first",
	ControllerEventsLast:         "controller.events.last",
	ControllerEventsCurrent:      "controller.events.current",
	ControllerEventsStatus:       "controller.events.status",
	ControllerDoor1:              "controller.door.1",
	ControllerDoor2:              "controller.door.2",
	ControllerDoor3:              "controller.door.3",
	ControllerDoor4:              "controller.door.4",
}
