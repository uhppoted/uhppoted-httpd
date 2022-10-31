package users

import (
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
)

const UserStatus = schema.Status
const UserCreated = schema.Created
const UserDeleted = schema.Deleted
const UserModified = schema.Modified
const UserName = schema.UserName
const UserUID = schema.UserUID
const UserRole = schema.UserRole
const UserPassword = schema.UserPassword
const UserOTP = schema.UserOTP

var lookup = map[schema.Suffix]string{
	UserStatus:   "user.status",
	UserCreated:  "user.created",
	UserDeleted:  "user.deleted",
	UserModified: "user.modified",
	UserName:     "user.name",
	UserUID:      "user.uid",
	UserRole:     "user.role",
	UserPassword: "user.password",
	UserOTP:      "user.otp",
}
