package groups

import ()

type info struct {
	Group       string `json:"group"`
	FieldName   string `json:"field"`
	Description string `json:"description"`
}

func (i info) ID() string {
	return ""
}

func (i info) Name() string {
	return i.Group
}

func (i info) Field() string {
	return i.FieldName
}

func (i info) Details() string {
	return i.Description
}
