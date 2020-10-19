package db

import (
	"github.com/uhppoted/uhppoted-httpd/types"
)

type DB interface {
	Groups() types.Groups
	CardHolders() ([]*CardHolder, error)
	Post(string, map[string]interface{}) (interface{}, error)

	ACL() ([]types.Permissions, error)
}

type CardHolder struct {
	ID     string
	Name   *types.Name
	Card   *types.Card
	From   *types.Date
	To     *types.Date
	Groups []*Permission
}

type Permission struct {
	ID    string
	GID   string
	Value bool
}

func (c *CardHolder) Copy() *CardHolder {
	name := c.Name.Copy()
	card := c.Card.Copy()
	var groups = make([]*Permission, len(c.Groups))

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
