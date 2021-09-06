package cards

import (
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type Cards interface {
	AsObjects() []interface{}
	Clone() Cards
	UpdateByOID(auth auth.OpAuth, oid string, value string) ([]interface{}, error)
	Print()

	CardHolders() CardHolders
	ACL() ([]types.Permissions, error)
	//	Post(map[string]interface{}, auth.OpAuth) (interface{}, error)
}
