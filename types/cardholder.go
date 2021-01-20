package types

import (
	"fmt"
)

type CardHolders map[string]*CardHolder

type CardHolder struct {
	ID     string
	Name   *Name
	Card   *Card
	From   *Date
	To     *Date
	Groups map[string]bool
}

func (c *CardHolder) Clone() *CardHolder {
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

func (c *CardHolder) AsRuleEntity() interface{} {
	type entity struct {
		Name   string
		Card   uint32
		Groups []string
	}

	if c != nil {
		cardNumber := uint32(0)
		if c.Card != nil {
			cardNumber = uint32(*c.Card)
		}

		groups := []string{}
		for k, v := range c.Groups {
			if v {
				groups = append(groups, k)
			}
		}

		return &entity{
			Name:   fmt.Sprintf("%v", c.Name),
			Card:   cardNumber,
			Groups: groups,
		}
	}

	return &entity{}
}
