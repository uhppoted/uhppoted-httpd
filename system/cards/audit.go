package cards

import (
	"fmt"
)

type info struct {
	Card      string `json:"card"`
	CardName  string `json:"name"`
	FieldName string `json:"field"`
	Current   string `json:"current"`
	Updated   string `json:"new"`
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
	return fmt.Sprintf("from '%v' to '%v'", i.Current, i.Updated)
}
