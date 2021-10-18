package doors

import (
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
)

type info struct {
	OID       catalog.OID `json:"OID"`
	Door      string      `json:"door"`
	FieldName string      `json:"field"`
	Current   string      `json:"current"`
	Updated   string      `json:"new"`
}

func (i info) Field() string {
	return i.FieldName
}
