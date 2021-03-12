package types

import (
	"encoding/json"
	"fmt"
	"net"
	"regexp"
)

type Address net.UDPAddr

const DEFAULT_PORT = 60000

func (a *Address) String() string {
	if a != nil {
		if a.Port == DEFAULT_PORT {
			return a.IP.String()
		} else {
			return (*net.UDPAddr)(a).String()
		}
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

	addr, err := Resolve(s)
	if err != nil {
		return err
	}

	*a = *addr

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
	if matched, err := regexp.MatchString(`[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}:[0-9]{1,5}`, s); err != nil {
		return nil, err
	} else if matched {
		if addr, err := net.ResolveUDPAddr("udp", s); err != nil {
			return nil, err
		} else {
			a := Address(*addr)
			return &a, nil
		}
	}

	if matched, err := regexp.MatchString(`[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}`, s); err != nil {
		return nil, err
	} else if matched {
		if ip := net.ParseIP(s); ip != nil {
			addr := Address(net.UDPAddr{
				IP:   ip,
				Port: DEFAULT_PORT,
				Zone: "",
			})

			return &addr, nil
		}
	}

	return nil, fmt.Errorf("%s is not a valid UDP address:port", s)
}
