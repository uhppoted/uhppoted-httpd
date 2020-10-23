package db

import (
	"github.com/uhppoted/uhppoted-httpd/types"
)

type DB interface {
	Groups() types.Groups
	CardHolders() (types.CardHolders, error)

	ACL() ([]types.Permissions, error)
	Post(map[string]interface{}) (interface{}, error)
}
