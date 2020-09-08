package db

import (
	"time"
)

type DB interface {
	Groups() []Group
	CardHolders() ([]*CardHolder, error)
	Update(map[string]interface{}) (interface{}, error)
}

type ID interface {
}

type Group struct {
	ID   string
	Name string
}

type CardHolder struct {
	ID         string
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
