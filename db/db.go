package db

import (
	"github.com/uhppoted/uhppoted-httpd/types"
)

type DB interface {
	Groups() []*Group
	CardHolders() ([]*CardHolder, error)
	Update(map[string]interface{}) (interface{}, error)

	ACL() ([]types.Permissions, error)
}

type CardHolder struct {
	ID     string
	Name   Name
	Card   Card
	From   types.Date
	To     types.Date
	Groups []*Permission
}

type Name struct {
	ID   string
	Name string
}

type Card struct {
	ID     string
	Number uint32
}

type Group struct {
	ID    string
	Name  string
	Doors []string
}

type Permission struct {
	ID    string
	GID   string
	Value bool
}

func (g *Group) Copy() *Group {
	replicant := Group{
		ID:    g.ID,
		Name:  g.Name,
		Doors: make([]string, len(g.Doors)),
	}

	copy(replicant.Doors, g.Doors)

	return &replicant
}

func (c *CardHolder) Copy() *CardHolder {
	replicant := &CardHolder{
		ID: c.ID,
		Name: Name{
			ID:   c.Name.ID,
			Name: c.Name.Name,
		},
		Card: Card{
			ID:     c.Card.ID,
			Number: c.Card.Number,
		},
		From:   c.From,
		To:     c.To,
		Groups: make([]*Permission, len(c.Groups)),
	}

	for i, g := range c.Groups {
		replicant.Groups[i] = g.Copy()
	}

	return replicant
}

func (p *Permission) Copy() *Permission {
	return &Permission{
		ID:    p.ID,
		GID:   p.GID,
		Value: p.Value,
	}
}
