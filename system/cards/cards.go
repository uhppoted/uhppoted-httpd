package cards

import (
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type Cards interface {
	Load(file string) error
	Save() error
	Print()
	Clone() Cards

	AsObjects() []interface{}
	UpdateByOID(auth auth.OpAuth, oid string, value string) ([]interface{}, error)
	Validate() error

	ACL(rules IRules) ([]types.Permissions, error)
}
