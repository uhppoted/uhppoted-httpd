package types

import (
	"encoding/json"
	"fmt"
	"net"
	"regexp"
)

type BindAddress net.UDPAddr

const BIND_PORT = 0

func (a *BindAddress) String() string {
	if a != nil {
		if a.Port == BIND_PORT {
			return a.IP.String()
		} else {
			return (*net.UDPAddr)(a).String()
		}
	}

	return ""
}

func (a *BindAddress) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
}

func (a *BindAddress) UnmarshalJSON(bytes []byte) error {
	var s string

	if err := json.Unmarshal(bytes, &s); err != nil {
		return err
	}

	addr, err := ResolveBindAddress(s)
	if err != nil {
		return err
	}

	*a = *addr

	return nil
}

func (a *BindAddress) Equal(addr *Address) bool {
	switch {
	case a == nil && addr == nil:
		return true

	case a != nil && addr != nil:
		return a.IP.Equal(addr.IP)

	default:
		return false
	}
}

func (a *BindAddress) Clone() *BindAddress {
	if a != nil {
		addr := *a
		return &addr
	}

	return nil
}

func ResolveBindAddress(s string) (*BindAddress, error) {
	if matched, err := regexp.MatchString(`[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}:[0-9]{1,5}`, s); err != nil {
		return nil, err
	} else if matched {
		if addr, err := net.ResolveUDPAddr("udp", s); err != nil {
			return nil, err
		} else {
			a := BindAddress(*addr)
			return &a, nil
		}
	}

	if matched, err := regexp.MatchString(`[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}`, s); err != nil {
		return nil, err
	} else if matched {
		if ip := net.ParseIP(s); ip != nil {
			addr := BindAddress(net.UDPAddr{
				IP:   ip,
				Port: BIND_PORT,
				Zone: "",
			})

			return &addr, nil
		}
	}

	return nil, fmt.Errorf("%s is not a valid UDP address:port", s)
}
