package groups

import (
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
)

const GroupStatus = schema.Status
const GroupCreated = schema.Created
const GroupDeleted = schema.Deleted
const GroupModified = schema.Modified
const GroupName = schema.GroupName
const GroupDoors = schema.GroupDoors

const DoorName = schema.DoorName

var lookup = map[schema.Suffix]string{
	GroupStatus:   "group.status",
	GroupCreated:  "group.created",
	GroupDeleted:  "group.deleted",
	GroupModified: "group.modified",
	GroupName:     "group.name",
	GroupDoors:    "group.doors",
}
