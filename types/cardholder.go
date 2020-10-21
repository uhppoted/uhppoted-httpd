package types

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
