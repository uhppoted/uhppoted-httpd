package groups

import (
	"fmt"
)

type info struct {
	OID       string `json:"OID"`
	Group     string `json:"group"`
	FieldName string `json:"field"`
	Current   string `json:"current"`
	Updated   string `json:"new"`
}

func (i info) Field() string {
	return i.FieldName
}

func (i info) Details() string {
	return fmt.Sprintf("from '%v' to '%v'", i.Current, i.Updated)
}
