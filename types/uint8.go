package types

import (
	"fmt"
)

type Uint8 uint8

func (u Uint8) String() string {
	if u > 0 {
		return fmt.Sprintf("%v", uint8(u))
	}

	return ""
}
