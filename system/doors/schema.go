package doors

import (
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
)

const DoorStatus = schema.Status
const DoorCreated = schema.Created
const DoorDeleted = schema.Deleted

const DoorName = schema.DoorName
const DoorDelay = schema.DoorDelay
const DoorDelayStatus = schema.DoorDelayStatus
const DoorDelayConfigured = schema.DoorDelayConfigured
const DoorDelayError = schema.DoorDelayError
const DoorDelayModified = schema.DoorDelayModified
const DoorControl = schema.DoorControl
const DoorControlStatus = schema.DoorControlStatus
const DoorControlConfigured = schema.DoorControlConfigured
const DoorControlError = schema.DoorControlError
const DoorControlModified = schema.DoorControlModified
const DoorKeypad = schema.DoorKeypad

var lookup = map[schema.Suffix]string{
	DoorStatus:            "door.status",
	DoorCreated:           "door.created",
	DoorDeleted:           "door.deleted",
	DoorName:              "door.name",
	DoorDelay:             "door.delay",
	DoorDelayStatus:       "door.delay.status",
	DoorDelayConfigured:   "door.delay.configured",
	DoorDelayError:        "door.delay.error",
	DoorDelayModified:     "door.delay.modified",
	DoorControl:           "door.control",
	DoorControlStatus:     "door.control.status",
	DoorControlConfigured: "door.control.configured",
	DoorControlError:      "door.control.error",
	DoorControlModified:   "door.control.modified",
	DoorKeypad:            "door.keypad",
}
