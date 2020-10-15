package db

import (
	"github.com/uhppoted/uhppoted-httpd/types"
)

type DB interface {
	Groups() []*Group
	CardHolders() ([]*CardHolder, error)
	Post(string, map[string]interface{}) (interface{}, error)

	ACL() ([]types.Permissions, error)
}

type CardHolder struct {
	ID     string
	Name   *Name
	Card   *Card
	From   *types.Date
	To     *types.Date
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
	var name *Name
	var card *Card
	var groups = make([]*Permission, len(c.Groups))

	if c.Name != nil {
		name = &Name{
			ID:   c.Name.ID,
			Name: c.Name.Name,
		}
	}

	if c.Card != nil {
		card = &Card{
			ID:     c.Card.ID,
			Number: c.Card.Number,
		}
	}

	for i, g := range c.Groups {
		groups[i] = g.Copy()
	}

	replicant := &CardHolder{
		ID:     c.ID,
		Name:   name,
		Card:   card,
		From:   c.From,
		To:     c.To,
		Groups: groups,
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
