package types

import (
	"fmt"
)

type Card uint32

func (c *Card) String() string {
	if c == nil {
		return ""
	}

	return fmt.Sprintf("%v", *c)
}

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

func (c *Card) Equals(card *Card) bool {
	if c == nil && card == nil {
		return true
	}

	if c != nil && card != nil {
		return uint32(*c) == uint32(*card)
	}

	return false
}
