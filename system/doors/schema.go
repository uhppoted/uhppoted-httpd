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
	DoorStatus:            "status",
	DoorCreated:           "created",
	DoorDeleted:           "deleted",
	DoorName:              "name",
	DoorDelay:             "delay",
	DoorDelayStatus:       "delay.status",
	DoorDelayConfigured:   "delay.configured",
	DoorDelayError:        "delay.error",
	DoorDelayModified:     "delay.modified",
	DoorControl:           "control",
	DoorControlStatus:     "control.status",
	DoorControlConfigured: "control.configured",
	DoorControlError:      "control.error",
	DoorControlModified:   "control.modified",
}
