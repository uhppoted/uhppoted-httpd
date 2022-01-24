package types

import (
	"encoding/json"
	"fmt"
)

type Uint8 uint8

func (u Uint8) String() string {
	if u > 0 {
		return fmt.Sprintf("%v", uint8(u))
	}

	return ""
}

func (u Uint8) MarshalJSON() ([]byte, error) {
	if u == 0 {
		return json.Marshal("")
	}

	return json.Marshal(uint8(u))
}

type Uint32 uint32

func (u Uint32) String() string {
	if u > 0 {
		return fmt.Sprintf("%v", uint32(u))
	}

	return ""
}

func (u Uint32) MarshalJSON() ([]byte, error) {
	if u == 0 {
		return json.Marshal("")
	}

	return json.Marshal(uint32(u))
}
