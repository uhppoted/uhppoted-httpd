package db

import (
	"time"
)

type DB interface {
	Groups() []Group
	CardHolders() []*CardHolder
	Update(map[string]interface{}) error
}

type ID interface {
}

type Group struct {
	ID   uint32
	Name string
}

type CardHolder struct {
	ID         uint32
	Name       string
	CardNumber uint32
	From       *time.Time
	To         *time.Time
	Groups     []*BoolVar
}

type BoolVar struct {
	ID    string
	Value bool
}
