package cards

import (
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type Cards interface {
	Groups() types.Groups
	CardHolders() CardHolders
	AsObjects() []interface{}

	ACL() ([]types.Permissions, error)
	Post(map[string]interface{}, auth.OpAuth) (interface{}, error)
}
