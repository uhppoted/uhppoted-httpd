package interfaces

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/uhppoted/uhppoted-httpd/system/catalog"
)

type Interfaces struct {
	LANs map[catalog.OID]*LANx
}

func NewInterfaces() Interfaces {
	return Interfaces{
		LANs: map[catalog.OID]*LANx{},
	}
}

func (ii *Interfaces) Load(blob json.RawMessage) error {
	rs := []json.RawMessage{}
	if err := json.Unmarshal(blob, &rs); err != nil {
		return err
	}

	for _, v := range rs {
		var l LANx
		if err := l.deserialize(v); err == nil {
			if _, ok := ii.LANs[l.OID]; ok {
				return fmt.Errorf("card '%v': duplicate OID (%v)", l.Name, l.OID)
			}

			ii.LANs[l.OID] = &l
		}
	}

	for _, v := range ii.LANs {
		catalog.PutInterface(v.OID)
	}

	return nil
}

func (ii Interfaces) Save() (json.RawMessage, error) {
	if err := validate(ii); err != nil {
		return nil, err
	}

	if err := scrub(ii); err != nil {
		return nil, err
	}

	serializable := []json.RawMessage{}
	for _, l := range ii.LANs {
		if l.IsValid() && !l.IsDeleted() {
			if record, err := l.serialize(); err == nil && record != nil {
				serializable = append(serializable, record)
			}
		}
	}

	return json.MarshalIndent(serializable, "", "  ")
}

func validate(ii Interfaces) error {
	names := map[string]string{}

	for k, l := range ii.LANs {
		if l.Deleted != nil {
			continue
		}

		if l.OID == "" {
			return fmt.Errorf("Invalid LAN OID (%v)", l.OID)
		}

		if k != l.OID {
			return fmt.Errorf("LAN %s: mismatched LAN OID %v (expected %v)", l.Name, l.OID, k)
		}

		n := strings.TrimSpace(strings.ToLower(l.Name))
		if v, ok := names[n]; ok && n != "" {
			return fmt.Errorf("'%v': duplicate LAN name (%v)", l.Name, v)
		}

		names[n] = l.Name
	}

	return nil
}

func scrub(ii Interfaces) error {
	return nil
}

func warn(err error) {
	log.Printf("ERROR %v", err)
}
