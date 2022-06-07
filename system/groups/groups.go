package groups

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	"github.com/uhppoted/uhppoted-httpd/system/db"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type Groups struct {
	groups map[schema.OID]Group
}

var guard sync.RWMutex

func NewGroups() Groups {
	return Groups{
		groups: map[schema.OID]Group{},
	}
}

func (gg Groups) Doors(groups ...schema.OID) []schema.OID {
	doors := []schema.OID{}

	for _, oid := range groups {
		if g, ok := gg.groups[oid]; ok {
			for k, v := range g.Doors {
				if v {
					doors = append(doors, k)
				}
			}
		}
	}

	return doors
}

func (gg *Groups) AsObjects(a *auth.Authorizator) []schema.Object {
	guard.RLock()
	defer guard.RUnlock()

	objects := []schema.Object{}

	for _, g := range gg.groups {
		if g.IsValid() || g.IsDeleted() {
			catalog.Join(&objects, g.AsObjects(a)...)
		}
	}

	return objects
}

func (gg *Groups) Create(a *auth.Authorizator, oid schema.OID, value string, dbc db.DBC) ([]schema.Object, error) {
	objects := []schema.Object{}

	if gg != nil {
		if g, err := gg.add(a, Group{}); err != nil {
			return nil, err
		} else if g == nil {
			return nil, fmt.Errorf("Failed to add 'new' group")
		} else {
			g.log(dbc, auth.UID(a), "add", "group", "", "", "Added 'new' group")

			catalog.Join(&objects, catalog.NewObject(g.OID, "new"))
			catalog.Join(&objects, catalog.NewObject2(g.OID, GroupCreated, g.created))
		}
	}

	return objects, nil
}

func (gg *Groups) Update(a *auth.Authorizator, oid schema.OID, value string, dbc db.DBC) ([]schema.Object, error) {
	objects := []schema.Object{}

	if gg != nil {
		for k, g := range gg.groups {
			if g.OID.Contains(oid) {
				objects, err := g.set(a, oid, value, dbc)
				if err == nil {
					gg.groups[k] = g
				}

				return objects, err
			}
		}
	}

	return objects, nil
}

func (gg *Groups) Delete(auth *auth.Authorizator, oid schema.OID, dbc db.DBC) ([]schema.Object, error) {
	if gg != nil {
		for k, g := range gg.groups {
			if g.OID == oid {
				objects, err := g.delete(auth, dbc)
				if err == nil {
					gg.groups[k] = g
				}

				return objects, err
			}
		}
	}

	return []schema.Object{}, nil
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
		catalog.PutT(g.CatalogGroup)
		catalog.PutV(g.OID, GroupName, g.Name)
		catalog.PutV(g.OID, GroupCreated, g.created)
	}

	return nil
}

func (gg Groups) Save() (json.RawMessage, error) {
	if err := gg.Validate(); err != nil {
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

func (gg *Groups) Group(oid schema.OID) (Group, bool) {
	g, ok := gg.groups[oid]

	return g, ok
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

// NTS: 'added' is specifically not cloned - it has a lifetime for the duration of
//      the 'shadow' copy only
// NTS: 'added' is specifically not cloned - it has a lifetime for the duration of
//      the 'shadow' copy only
func (gg *Groups) Clone() Groups {
	guard.RLock()
	defer guard.RUnlock()

	shadow := Groups{
		groups: map[schema.OID]Group{},
	}

	for k, v := range gg.groups {
		shadow.groups[k] = v.clone()
	}

	return shadow
}

func (gg Groups) Validate() error {
	names := map[string]string{}

	for k, g := range gg.groups {
		if g.IsDeleted() {
			continue
		}

		if g.OID == "" {
			return fmt.Errorf("Invalid group OID (%v)", g.OID)
		} else if k != g.OID {
			return fmt.Errorf("Group %s: mismatched group OID %v (expected %v)", g.Name, g.OID, k)
		}

		if err := g.validate(); err != nil {
			if !g.modified.IsZero() {
				return err
			}
		}

		n := strings.TrimSpace(strings.ToLower(g.Name))
		if v, ok := names[n]; ok && n != "" {
			return fmt.Errorf("'%v': duplicate group name (%v)", g.Name, v)
		}

		names[n] = g.Name
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
	oid := catalog.NewT(g.CatalogGroup)
	if _, ok := gg.groups[oid]; ok {
		return nil, fmt.Errorf("catalog returned duplicate OID (%v)", oid)
	}

	group := g.clone()
	group.OID = oid
	group.created = types.TimestampNow()

	if a != nil {
		if err := a.CanAdd(&group, auth.Groups); err != nil {
			return nil, err
		}
	}

	gg.groups[group.OID] = group

	return &group, nil
}
