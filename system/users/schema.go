package users

import (
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
)

const UserStatus = catalog.Status
const UserCreated = catalog.Created
const UserDeleted = catalog.Deleted
const UserModified = catalog.Modified
const UserName = catalog.UserName
const UserUID = catalog.UserUID
const UserRole = catalog.UserRole
const UserPassword = catalog.UserPassword

var lookup = map[catalog.Suffix]string{
	UserStatus:   "user.status",
	UserCreated:  "user.created",
	UserDeleted:  "user.deleted",
	UserModified: "user.modified",
	UserName:     "user.name",
	UserUID:      "user.uid",
	UserRole:     "user.role",
	UserPassword: "user.password",
}
