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

func (n *Name) String() string {
	if n == nil {
		return ""
	}

	return string(*n)
}

func (n *Name) IsValid() bool {
	if n != nil && strings.TrimSpace(string(*n)) != "" {
		return true
	}

	return false
}

func (n *Name) Equals(name *Name) bool {
	if n == nil && name == nil {
		return true
	}

	if n != nil && name != nil {
		return string(*n) == string(*name)
	}

	return false
}
