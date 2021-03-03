package types

import (
	"encoding/json"
	"fmt"
	"net"
	"regexp"
)

type Address net.UDPAddr

func (a *Address) String() string {
	if a != nil {
		return (*net.UDPAddr)(a).String()
	}

	return ""
}

func (a *Address) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
}

func (a *Address) UnmarshalJSON(bytes []byte) error {
	var s string

	if err := json.Unmarshal(bytes, &s); err != nil {
		return err
	}

	addr, err := net.ResolveUDPAddr("udp", s)
	if err != nil {
		return err
	}

	*a = Address(*addr)

	return nil
}

func (a *Address) Equal(addr *Address) bool {
	switch {
	case a == nil && addr == nil:
		return true

	case a != nil && addr != nil:
		return a.IP.Equal(addr.IP)

	default:
		return false
	}
}

func (a *Address) Clone() *Address {
	if a != nil {
		addr := *a
		return &addr
	}

	return nil
}

func Resolve(s string) (*Address, error) {
	matched, err := regexp.MatchString(`[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}:[0-9]{1,5}`, s)
	if err != nil {
		return nil, err
	}

	if !matched {
		return nil, fmt.Errorf("%s is not a valid UDP address:port", s)
	}

	addr, err := net.ResolveUDPAddr("udp", s)
	if err != nil {
		return nil, err
	}

	a := Address(*addr)

	return &a, nil
}
