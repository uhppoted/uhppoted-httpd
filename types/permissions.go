package types

import ()

type Permissions struct {
	CardNumber uint32
	From       Date
	To         Date
	Doors      []string
}
