package system

import (
	"encoding/json"
	"net"
)

type address net.UDPAddr

func (a *address) String() string {
	if a != nil {
		return (*net.UDPAddr)(a).String()
	}

	return ""
}

func (a *address) UnmarshalJSON(bytes []byte) error {
	var s string

	if err := json.Unmarshal(bytes, &s); err != nil {
		return err
	}

	addr, err := net.ResolveUDPAddr("udp", s)
	if err != nil {
		return err
	}

	*a = address(*addr)

	return nil
}

func (a *address) Equal(addr net.IP) bool {
	if a != nil {
		return a.IP.Equal(addr)
	}

	return false
}
