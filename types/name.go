package types

import (
	"strings"
)

type Name string

func (n *Name) Copy() *Name {
	if n == nil {
		return nil
	}

	name := *n

	return &name
}

func (n *Name) IsValid() bool {
	if n != nil && strings.TrimSpace(string(*n)) != "" {
		return true
	}

	return false
}
