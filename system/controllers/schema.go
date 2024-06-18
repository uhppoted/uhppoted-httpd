package controllers

import (
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
)

const LANStatus = schema.Status
const LANType = schema.Type
const LANName = schema.InterfaceName
const LANBindAddress = schema.LANBindAddress
const LANBroadcastAddress = schema.LANBroadcastAddress
const LANListenAddress = schema.LANListenAddress
const DoorDelay = schema.DoorDelay
const DoorDelayModified = schema.DoorDelayModified
const DoorControl = schema.DoorControl
const DoorControlModified = schema.DoorControlModified

const ControllerStatus = schema.Status
const ControllerCreated = schema.Created
const ControllerDeleted = schema.Deleted
const ControllerTouched = schema.Touched
const ControllerName = schema.ControllerName
const ControllerDeviceID = schema.ControllerDeviceID
const ControllerEndpoint = schema.ControllerEndpoint
const ControllerEndpointStatus = schema.ControllerEndpointStatus
const ControllerEndpointAddress = schema.ControllerEndpointAddress
const ControllerEndpointProtocol = schema.ControllerEndpointProtocol
const ControllerEndpointConfigured = schema.ControllerEndpointConfigured
const ControllerDateTime = schema.ControllerDateTime
const ControllerDateTimeStatus = schema.ControllerDateTimeStatus
const ControllerDateTimeCurrent = schema.ControllerDateTimeCurrent
const ControllerDateTimeConfigured = schema.ControllerDateTimeConfigured
const ControllerDateTimeModified = schema.ControllerDateTimeModified
const ControllerCardsStatus = schema.ControllerCardsStatus
const ControllerCardsCount = schema.ControllerCardsCount
const ControllerEventsFirst = schema.ControllerEventsFirst
const ControllerEventsLast = schema.ControllerEventsLast
const ControllerEventsCurrent = schema.ControllerEventsCurrent
const ControllerEventsStatus = schema.ControllerEventsStatus
const ControllerDoor1 = schema.ControllerDoor1
const ControllerDoor2 = schema.ControllerDoor2
const ControllerDoor3 = schema.ControllerDoor3
const ControllerDoor4 = schema.ControllerDoor4
const ControllerInterlock = schema.ControllerInterlock

var lookup = map[schema.Suffix]string{
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
	ControllerDateTimeConfigured: "controller.datetime.configured",
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
	ControllerInterlock:          "controller.interlock",
}
