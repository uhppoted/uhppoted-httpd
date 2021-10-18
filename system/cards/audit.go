package cards

import ()

type info struct {
	OID       string `json:"OID"`
	Card      string `json:"card"`
	FieldName string `json:"field"`
	Current   string `json:"current"`
	Updated   string `json:"new"`
}

func (i info) Field() string {
	return i.FieldName
}
