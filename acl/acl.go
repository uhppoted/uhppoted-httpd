package acl

import (
	"time"
)

type Date time.Time

type Permissions struct {
	CardNumber uint32
	From       *Date
	To         *Date
	Doors      []string
}

type ACL interface {
	Permissions() map[uint32]Permissions
}
