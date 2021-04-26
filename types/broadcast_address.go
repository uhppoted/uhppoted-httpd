package types

import (
	"encoding/json"
	"fmt"
	"net"
	"regexp"
)

type BroadcastAddress net.UDPAddr

const BROADCAST_PORT = 60000

func (a *BroadcastAddress) String() string {
	if a != nil {
		if a.Port == BROADCAST_PORT {
			return a.IP.String()
		} else {
			return (*net.UDPAddr)(a).String()
		}
	}

	return ""
}

func (a *BroadcastAddress) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
}

func (a *BroadcastAddress) UnmarshalJSON(bytes []byte) error {
	var s string

	if err := json.Unmarshal(bytes, &s); err != nil {
		return err
	}

	addr, err := ResolveBroadcastAddress(s)
	if err != nil {
		return err
	}

	*a = *addr

	return nil
}

func (a *BroadcastAddress) Equal(addr *BroadcastAddress) bool {
	switch {
	case a == nil && addr == nil:
		return true

	case a != nil && addr != nil:
		return a.IP.Equal(addr.IP)

	default:
		return false
	}
}

func (a *BroadcastAddress) Clone() *BroadcastAddress {
	if a != nil {
		addr := *a
		return &addr
	}

	return nil
}

func ResolveBroadcastAddress(s string) (*BroadcastAddress, error) {
	if matched, err := regexp.MatchString(`[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}:[0-9]{1,5}`, s); err != nil {
		return nil, err
	} else if matched {
		if addr, err := net.ResolveUDPAddr("udp", s); err != nil {
			return nil, err
		} else {
			a := BroadcastAddress(*addr)
			return &a, nil
		}
	}

	if matched, err := regexp.MatchString(`[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}`, s); err != nil {
		return nil, err
	} else if matched {
		if ip := net.ParseIP(s); ip != nil {
			addr := BroadcastAddress(net.UDPAddr{
				IP:   ip,
				Port: BROADCAST_PORT,
				Zone: "",
			})

			return &addr, nil
		}
	}

	return nil, fmt.Errorf("%s is not a valid UDP address:port", s)
}
