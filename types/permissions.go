package types

import (
	"fmt"
)

type Permissions struct {
	CardNumber uint32
	From       Date
	To         Date
	Doors      []string
}

func (p Permissions) String() string {
	return fmt.Sprintf("%-10v %v %v %v", p.CardNumber, p.From, p.To, p.Doors)
}
