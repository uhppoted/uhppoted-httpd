package doors

import (
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
)

const DoorStatus = catalog.Status
const DoorCreated = catalog.Created
const DoorDeleted = catalog.Deleted

const DoorName = catalog.DoorName
const DoorDelay = catalog.DoorDelay
const DoorDelayStatus = catalog.DoorDelayStatus
const DoorDelayConfigured = catalog.DoorDelayConfigured
const DoorDelayError = catalog.DoorDelayError
const DoorDelayModified = catalog.DoorDelayModified
const DoorControl = catalog.DoorControl
const DoorControlStatus = catalog.DoorControlStatus
const DoorControlConfigured = catalog.DoorControlConfigured
const DoorControlError = catalog.DoorControlError
const DoorControlModified = catalog.DoorControlModified

var lookup = map[catalog.Suffix]string{
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
}
