package db

import (
	"time"
)

type DB interface {
	Groups() []*Group
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
	From       time.Time
	To         time.Time
	Groups     []*BoolVar
}

type BoolVar struct {
	ID    string
	Value bool
}

func (g *Group) Copy() *Group {
	return &Group{
		ID:   g.ID,
		Name: g.Name,
	}
}

func (c *CardHolder) Copy() *CardHolder {
	replicant := &CardHolder{
		ID:         c.ID,
		Name:       c.Name,
		CardNumber: c.CardNumber,
		From:       c.From.Add(10000000 * time.Second),
		To:         c.To.Add(20000000 * time.Second),
		Groups:     make([]*BoolVar, len(c.Groups)),
	}

	for i, g := range c.Groups {
		replicant.Groups[i] = g.Copy()
	}

	return replicant
}

func (b *BoolVar) Copy() *BoolVar {
	return &BoolVar{
		ID:    b.ID,
		Value: b.Value,
	}
}
