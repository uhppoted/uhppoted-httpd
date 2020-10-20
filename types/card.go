package types

type Card uint32

func (c *Card) Copy() *Card {
	if c == nil {
		return nil
	}

	card := *c

	return &card
}

func (c *Card) IsValid() bool {
	if c != nil && *c != 0 {
		return true
	}

	return false
}