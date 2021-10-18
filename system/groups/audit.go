package groups

import ()

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
