package groups

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/db"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type Groups struct {
	groups map[catalog.OID]Group
}

var guard sync.RWMutex

func NewGroups() Groups {
	return Groups{
		groups: map[catalog.OID]Group{},
	}
}

func (gg *Groups) Group(oid catalog.OID) (Group, bool) {
	g, ok := gg.groups[oid]

	return g, ok
}

func (gg *Groups) Load(blob json.RawMessage) error {
	rs := []json.RawMessage{}
	if err := json.Unmarshal(blob, &rs); err != nil {
		return err
	}

	for _, v := range rs {
		var g Group
		if err := g.deserialize(v); err == nil {
			if _, ok := gg.groups[g.OID]; ok {
				return fmt.Errorf("group '%v': duplicate OID (%v)", g.Name, g.OID)
			}

			gg.groups[g.OID] = g
		}
	}

	for _, g := range gg.groups {
		catalog.PutGroup(g.OID)
		catalog.PutV(g.OID, GroupName, g.Name)
		catalog.PutV(g.OID, GroupCreated, g.created)
	}

	return nil
}

func (gg Groups) Save() (json.RawMessage, error) {
	if err := validate(gg); err != nil {
		return nil, err
	}

	if err := scrub(gg); err != nil {
		return nil, err
	}

	serializable := []json.RawMessage{}

	for _, g := range gg.groups {
		if g.IsValid() && !g.IsDeleted() {
			if record, err := g.serialize(); err == nil && record != nil {
				serializable = append(serializable, record)
			}
		}
	}

	return json.MarshalIndent(serializable, "", "  ")
}

func (gg Groups) Print() {
	serializable := []json.RawMessage{}
	for _, g := range gg.groups {
		if g.IsValid() && !g.IsDeleted() {
			if record, err := g.serialize(); err == nil && record != nil {
				serializable = append(serializable, record)
			}
		}
	}

	if b, err := json.MarshalIndent(serializable, "", "  "); err == nil {
		fmt.Printf("----------------- GROUPS\n%s\n", string(b))
	}
}

func (gg *Groups) Clone() Groups {
	guard.RLock()
	defer guard.RUnlock()

	shadow := Groups{
		groups: map[catalog.OID]Group{},
	}

	for k, v := range gg.groups {
		shadow.groups[k] = v.clone()
	}

	return shadow
}

func (gg *Groups) AsObjects(auth auth.OpAuth) catalog.Objects {
	guard.RLock()
	defer guard.RUnlock()

	objects := catalog.Objects{}

	for _, g := range gg.groups {
		if g.IsValid() || g.IsDeleted() {
			objects.Append(g.AsObjects(auth)...)
		}
	}

	return objects
}

func (gg *Groups) UpdateByOID(auth auth.OpAuth, oid catalog.OID, value string, dbc db.DBC) ([]catalog.Object, error) {
	if gg == nil {
		return nil, nil
	}

	for k, g := range gg.groups {
		if g.OID.Contains(oid) {
			objects, err := g.set(auth, oid, value, dbc)
			if err == nil {
				gg.groups[k] = g
			}

			return objects, err
		}
	}

	objects := []catalog.Object{}

	if oid == "<new>" {
		if g, err := gg.add(auth, Group{}); err != nil {
			return nil, err
		} else if g == nil {
			return nil, fmt.Errorf("Failed to add 'new' group")
		} else {
			g.log(auth, "add", g.OID, "group", "Added 'new' group", dbc)

			gg.groups[g.OID] = *g
			objects = append(objects, catalog.NewObject(g.OID, "new"))
			objects = append(objects, catalog.NewObject2(g.OID, GroupCreated, g.created))
		}
	}

	return objects, nil
}

func (gg *Groups) Validate() error {
	if gg != nil {
		return validate(*gg)
	}

	return nil
}

func (gg *Groups) Sweep(retention time.Duration) {
	if gg != nil {
		cutoff := time.Now().Add(-retention)
		for i, v := range gg.groups {
			if v.IsDeleted() && v.deleted.Before(cutoff) {
				delete(gg.groups, i)
			}
		}
	}
}

func (gg *Groups) add(a auth.OpAuth, g Group) (*Group, error) {
	oid := catalog.NewGroup()

	record := g.clone()
	record.OID = oid
	record.created = types.TimestampNow()

	if a != nil {
		if err := a.CanAdd(&record, auth.Groups); err != nil {
			return nil, err
		}
	}

	return &record, nil
}

func validate(gg Groups) error {
	names := map[string]string{}

	for k, g := range gg.groups {
		if g.IsDeleted() {
			continue
		}

		if g.OID == "" {
			return fmt.Errorf("Invalid group OID (%v)", g.OID)
		}

		if k != g.OID {
			return fmt.Errorf("Group %s: mismatched group OID %v (expected %v)", g.Name, g.OID, k)
		}

		n := strings.TrimSpace(strings.ToLower(g.Name))
		if v, ok := names[n]; ok && n != "" {
			return fmt.Errorf("'%v': duplicate group name (%v)", g.Name, v)
		}

		names[n] = g.Name
	}

	return nil
}

func scrub(gg Groups) error {
	return nil
}
