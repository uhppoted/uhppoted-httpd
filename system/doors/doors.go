package doors

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	"github.com/uhppoted/uhppoted-httpd/system/db"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type Doors struct {
	doors map[schema.OID]Door
	file  string
}

type object schema.Object

var guard sync.RWMutex

func NewDoors() Doors {
	return Doors{
		doors: map[schema.OID]Door{},
	}
}

func (dd *Doors) Door(oid schema.OID) (Door, bool) {
	d, ok := dd.doors[oid]

	return d, ok
}

func (dd *Doors) AsObjects(a *auth.Authorizator) []schema.Object {
	objects := []schema.Object{}

	for _, d := range dd.doors {
		if d.IsValid() || d.IsDeleted() {
			catalog.Join(&objects, d.AsObjects(a)...)
		}
	}

	return objects
}

func (dd *Doors) Load(blob json.RawMessage) error {
	rs := []json.RawMessage{}
	if err := json.Unmarshal(blob, &rs); err != nil {
		return err
	}

	for _, v := range rs {
		var d Door
		if err := d.deserialize(v); err != nil {
			warn(err)
		} else {
			if _, ok := dd.doors[d.OID]; ok {
				return fmt.Errorf("door '%v': duplicate OID (%v)", d.Name, d.OID)
			}

			dd.doors[d.OID] = d
		}
	}

	for _, d := range dd.doors {
		catalog.PutT(d.CatalogDoor)
		catalog.PutV(d.OID, DoorName, d.Name)
		catalog.PutV(d.OID, DoorDelayConfigured, d.delay)
		catalog.PutV(d.OID, DoorDelayModified, false)
		catalog.PutV(d.OID, DoorControlConfigured, d.mode)
		catalog.PutV(d.OID, DoorControlModified, false)
	}

	return nil
}

func (dd Doors) Save() (json.RawMessage, error) {
	if err := dd.Validate(); err != nil {
		return nil, err
	}

	serializable := []json.RawMessage{}
	for _, d := range dd.doors {
		if d.IsValid() && !d.IsDeleted() {
			if record, err := d.serialize(); err == nil && record != nil {
				serializable = append(serializable, record)
			}
		}
	}

	return json.MarshalIndent(serializable, "", "  ")
}

func (dd *Doors) Sweep(retention time.Duration) {
	if dd != nil {
		cutoff := time.Now().Add(-retention)
		for i, v := range dd.doors {
			if v.IsDeleted() && v.deleted.Before(cutoff) {
				delete(dd.doors, i)
			}
		}
	}
}

func (dd *Doors) ByName(name string) (Door, bool) {
	clean := func(s string) string {
		return strings.ToLower(regexp.MustCompile(`\s+`).ReplaceAllString(s, ""))
	}

	for _, d := range dd.doors {
		p := clean(d.Name)
		q := clean(name)

		if p == q {
			return d, false
		}
	}

	return Door{}, false
}

func (dd Doors) Print() {
	serializable := []json.RawMessage{}
	for _, d := range dd.doors {
		if d.IsValid() && !d.IsDeleted() {
			if record, err := d.serialize(); err == nil && record != nil {
				serializable = append(serializable, record)
			}
		}
	}

	if b, err := json.MarshalIndent(serializable, "", "  "); err == nil {
		fmt.Printf("----------------- DOORS\n%s\n", string(b))
	}
}

func (dd *Doors) UpdateByOID(a *auth.Authorizator, oid schema.OID, value string, dbc db.DBC) ([]schema.Object, error) {
	objects := []schema.Object{}

	if dd != nil {
		for k, d := range dd.doors {
			if d.OID.Contains(oid) {
				objects, err := d.set(a, oid, value, dbc)
				if err == nil {
					dd.doors[k] = d
				}

				return objects, err
			}
		}

		if oid == "<new>" {
			if d, err := dd.add(a, Door{}); err != nil {
				return nil, err
			} else if d == nil {
				return nil, fmt.Errorf("Failed to add 'new' door")
			} else {
				d.log(auth.UID(a), "add", d.OID, "door", fmt.Sprintf("Added 'new' door"), dbc)

				catalog.Join(&objects, catalog.NewObject(d.OID, "new"))
				catalog.Join(&objects, catalog.NewObject2(d.OID, DoorCreated, d.created))
			}
		}
	}

	return objects, nil
}

func (dd *Doors) DeleteByOID(a *auth.Authorizator, oid schema.OID, dbc db.DBC) ([]schema.Object, error) {
	objects := []schema.Object{}

	if dd != nil {
		for k, d := range dd.doors {
			if d.OID == oid {
				objects, err := d.delete(a, dbc)
				if err == nil {
					dd.doors[k] = d
				}

				return objects, err
			}
		}
	}

	return objects, nil
}

func (dd *Doors) add(a auth.OpAuth, d Door) (*Door, error) {
	oid := catalog.NewT(d.CatalogDoor)
	if _, ok := dd.doors[oid]; ok {
		return nil, fmt.Errorf("catalog returned duplicate OID (%v)", oid)
	}

	door := d.clone()
	door.OID = oid
	door.created = types.TimestampNow()

	if a != nil {
		if err := a.CanAdd(&door, auth.Doors); err != nil {
			return nil, err
		}
	}

	dd.doors[door.OID] = door

	return &door, nil
}

// NTS: 'added' is specifically not cloned - it has a lifetime for the duration of
//      the 'shadow' copy only
func (dd *Doors) Clone() Doors {
	guard.RLock()
	defer guard.RUnlock()

	shadow := Doors{
		doors: map[schema.OID]Door{},
		file:  dd.file,
	}

	for k, v := range dd.doors {
		shadow.doors[k] = v.clone()
	}

	return shadow
}

func (dd Doors) Validate() error {
	names := map[string]string{}

	for k, d := range dd.doors {
		if d.IsDeleted() {
			continue
		}

		if d.OID == "" {
			return fmt.Errorf("Invalid door OID (%v)", d.OID)
		} else if k != d.OID {
			return fmt.Errorf("Door %s: mismatched door OID %v (expected %v)", d.Name, d.OID, k)
		}

		if err := d.validate(); err != nil {
			if !d.modified.IsZero() {
				return err
			}
		}

		n := strings.TrimSpace(strings.ToLower(d.Name))
		if v, ok := names[n]; ok && n != "" {
			return fmt.Errorf("'%v': duplicate door name (%v)", d.Name, v)
		}

		names[n] = d.Name
	}

	return nil
}

func warn(err error) {
	log.Printf("ERROR %v", err)
}
