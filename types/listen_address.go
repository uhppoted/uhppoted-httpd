package types

import (
	"encoding/json"
	"fmt"
	"net"
	"regexp"
)

type ListenAddress net.UDPAddr

func (a *ListenAddress) String() string {
	if a != nil {
		return (*net.UDPAddr)(a).String()
	}

	return ""
}

func (a *ListenAddress) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
}

func (a *ListenAddress) UnmarshalJSON(bytes []byte) error {
	var s string

	if err := json.Unmarshal(bytes, &s); err != nil {
		return err
	}

	addr, err := ResolveListenAddress(s)
	if err != nil {
		return err
	}

	*a = *addr

	return nil
}

func (a *ListenAddress) Equal(addr *ListenAddress) bool {
	switch {
	case a == nil && addr == nil:
		return true

	case a != nil && addr != nil:
		return a.IP.Equal(addr.IP)

	default:
		return false
	}
}

func (a *ListenAddress) Clone() *ListenAddress {
	if a != nil {
		addr := *a
		return &addr
	}

	return nil
}

func ResolveListenAddress(s string) (*ListenAddress, error) {
	if matched, err := regexp.MatchString(`[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}:[0-9]{1,5}`, s); err != nil {
		return nil, err
	} else if matched {
		if addr, err := net.ResolveUDPAddr("udp", s); err != nil {
			return nil, err
		} else if addr.Port == 0 || addr.Port == 60000 {
			return nil, fmt.Errorf("%v: invalid 'listen' port (%v)", addr,addr.Port)
		} else {
			a := ListenAddress(*addr)
			return &a, nil
		}
	}

	return nil, fmt.Errorf("%s is not a valid UDP address:port", s)
}
