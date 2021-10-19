package cards

import ()

type info struct {
	Card        string `json:"card"`
	CardName    string `json:"name"`
	FieldName   string `json:"field"`
	Description string `json:"description"`
}

func (i info) ID() string {
	return i.Card
}

func (i info) Name() string {
	return i.CardName
}

func (i info) Field() string {
	return i.FieldName
}

func (i info) Details() string {
	return i.Description
}
