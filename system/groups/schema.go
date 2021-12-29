package groups

import (
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
)

const GroupStatus = catalog.Status
const GroupCreated = catalog.Created
const GroupDeleted = catalog.Deleted
const GroupModified = catalog.Modified
const GroupName = catalog.GroupName
const GroupDoors = catalog.GroupDoors

const DoorName = catalog.DoorName

var lookup = map[catalog.Suffix]string{
	GroupStatus:   "group.status",
	GroupCreated:  "group.created",
	GroupDeleted:  "group.deleted",
	GroupModified: "group.modified",
	GroupName:     "group.name",
	GroupDoors:    "group.doors",
}
