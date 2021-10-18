package doors

import (
	"fmt"

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

func (i info) Details() string {
	return fmt.Sprintf("from '%v' to '%v'", i.Current, i.Updated)
}
