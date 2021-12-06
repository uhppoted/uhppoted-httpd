package interfaces

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/db"
)

type Interfaces struct {
	LANs map[catalog.OID]*LANx
}

const BLANK = "'blank'"

var guard sync.RWMutex

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

func (ii *Interfaces) Clone() Interfaces {
	guard.RLock()
	defer guard.RUnlock()

	shadow := Interfaces{
		LANs: map[catalog.OID]*LANx{},
	}

	for k, v := range ii.LANs {
		clone := v.Clone()
		shadow.LANs[k] = &clone
	}

	return shadow
}

func (ii *Interfaces) AsObjects() []interface{} {
	objects := []interface{}{}

	for _, l := range ii.LANs {
		if l.IsValid() {
			if v := l.AsObjects(); v != nil {
				objects = append(objects, v...)
			}
		}
	}

	return objects
}

func (ii *Interfaces) UpdateByOID(auth auth.OpAuth, oid catalog.OID, value string, dbc db.DBC) ([]catalog.Object, error) {
	if ii == nil {
		return nil, nil
	}

	for _, l := range ii.LANs {
		if l != nil && l.OID.Contains(oid) {
			return l.set(auth, oid, value, dbc)
		}
	}

	objects := []catalog.Object{}

	if oid == "<new>" {
		if l, err := ii.add(auth, LANx{}); err != nil {
			return nil, err
		} else if l == nil {
			return nil, fmt.Errorf("Failed to add 'new' interface")
		} else {
			l.log(auth, "add", l.OID, "interface", fmt.Sprintf("Added 'new' interface"), "", "", dbc)
			objects = append(objects, catalog.NewObject(l.OID, "new"))
			objects = append(objects, catalog.NewObject2(l.OID, LANStatus, "new"))
			objects = append(objects, catalog.NewObject2(l.OID, LANCreated, l.created))
		}
	}

	return objects, nil
}

func (ii Interfaces) Validate() error {
	return validate(ii)
}

func (ii *Interfaces) add(auth auth.OpAuth, l LANx) (*LANx, error) {
	return nil, fmt.Errorf("NOT SUPPORTED")
}

func validate(ii Interfaces) error {
	names := map[string]string{}

	for k, l := range ii.LANs {
		if l.deleted != nil {
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

func stringify(i interface{}, defval string) string {
	s := ""
	switch v := i.(type) {
	case *uint32:
		if v != nil {
			s = fmt.Sprintf("%v", *v)
		}

	case *string:
		if v != nil {
			s = fmt.Sprintf("%v", *v)
		}

	default:
		s = fmt.Sprintf("%v", i)
	}

	if s != "" {
		return s
	}

	return defval
}

func warn(err error) {
	log.Printf("ERROR %v", err)
}