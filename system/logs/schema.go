package logs

import (
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
)

const LogTimestamp = catalog.LogTimestamp
const LogUID = catalog.LogUID
const LogItem = catalog.LogItem
const LogItemID = catalog.LogItemID
const LogItemName = catalog.LogItemName
const LogField = catalog.LogField
const LogDetails = catalog.LogDetails

const ControllerName = catalog.ControllerName
const ControllerDeviceID = catalog.ControllerDeviceID

var lookup = map[catalog.Suffix]string{
	LogTimestamp: "log.timestamp",
	LogUID:       "log.UID",
	LogItem:      "log.item",
	LogItemID:    "log.item.ID",
	LogItemName:  "log.item.name",
	LogField:     "log.field",
	LogDetails:   "log.details",
}
