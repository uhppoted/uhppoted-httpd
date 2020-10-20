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
	Groups map[string]bool
}

func (c *CardHolder) Copy() *CardHolder {
	name := c.Name.Copy()
	card := c.Card.Copy()
	var groups = map[string]bool{}

	for gid, g := range c.Groups {
		groups[gid] = g
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
