package logs

import (
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
)

const LogsOID = schema.LogsOID
const LogsFirst = schema.LogsFirst
const LogsLast = schema.LogsLast

const LogTimestamp = schema.LogTimestamp
const LogUID = schema.LogUID
const LogItem = schema.LogItem
const LogItemID = schema.LogItemID
const LogItemName = schema.LogItemName
const LogField = schema.LogField
const LogDetails = schema.LogDetails

const ControllerName = schema.ControllerName
const ControllerDeviceID = schema.ControllerDeviceID

var lookup = map[schema.Suffix]string{
	LogTimestamp: "log.timestamp",
	LogUID:       "log.UID",
	LogItem:      "log.item",
	LogItemID:    "log.item.ID",
	LogItemName:  "log.item.name",
	LogField:     "log.field",
	LogDetails:   "log.details",
}
