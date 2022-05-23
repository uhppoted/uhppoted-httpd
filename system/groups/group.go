package groups

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	"github.com/uhppoted/uhppoted-httpd/system/db"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type Group struct {
	catalog.CatalogGroup
	Name  string              `json:"name"`
	Doors map[schema.OID]bool `json:"doors"`

	created  types.Timestamp
	modified types.Timestamp
	deleted  types.Timestamp
}

type kv = struct {
	field schema.Suffix
	value interface{}
}

const BLANK = "'blank'"

var created = types.TimestampNow()

func (g Group) String() string {
	return fmt.Sprintf("%v", g.Name)
}

func (g Group) IsValid() bool {
	return g.validate() == nil
}

func (g Group) validate() error {
	if strings.TrimSpace(g.Name) == "" {
		return fmt.Errorf("Group name is blank")
	}

	return nil
}

func (g Group) IsDeleted() bool {
	return !g.deleted.IsZero()
}

func (g *Group) AsObjects(a *auth.Authorizator) []schema.Object {
	list := []kv{}

	if g.IsDeleted() {
		list = append(list, kv{GroupDeleted, g.deleted})
	} else {
		name := g.Name

		list = append(list, kv{GroupStatus, g.Status()})
		list = append(list, kv{GroupCreated, g.created})
		list = append(list, kv{GroupDeleted, g.deleted})
		list = append(list, kv{GroupName, name})

		doors := catalog.GetDoors()
		re := regexp.MustCompile(`^(.*?)(\.[0-9]+)$`)

		for _, door := range doors {
			d := fmt.Sprintf("%v", door)

			if m := re.FindStringSubmatch(d); m != nil && len(m) > 2 {
				did := m[2]
				allowed := g.Doors[door]

				list = append(list, kv{GroupDoors.Append(did), allowed})
				list = append(list, kv{GroupDoors.Append(did + ".1"), door})
			}
		}
	}

	return g.toObjects(list, a)
}

func (g *Group) AsRuleEntity() (string, interface{}) {
	entity := struct {
		Name  string
		Doors map[string]bool
	}{
		Name:  "",
		Doors: map[string]bool{},
	}

	if g != nil {
		entity.Name = fmt.Sprintf("%v", g.Name)

		doors := catalog.GetDoors()
		for _, d := range doors {
			allowed := g.Doors[d]
			door := catalog.GetV(d, DoorName)

			if v := fmt.Sprintf("%v", door); v != "" {
				entity.Doors[v] = allowed
			}
		}
	}

	return "group", &entity
}

func (g Group) Status() types.Status {
	if g.IsDeleted() {
		return types.StatusDeleted
	}

	return types.StatusOk
}

func (g *Group) set(a *auth.Authorizator, oid schema.OID, value string, dbc db.DBC) ([]schema.Object, error) {
	if g == nil {
		return []schema.Object{}, nil
	}

	if g.IsDeleted() {
		return g.toObjects([]kv{{GroupDeleted, g.deleted}}, a), fmt.Errorf("Group has been deleted")
	}

	f := func(field string, value interface{}) error {
		if a != nil {
			return a.CanUpdate(g, field, value, auth.Groups)
		}

		return nil
	}

	uid := auth.UID(a)
	original := g.clone()
	list := []kv{}

	switch {
	case oid == g.OID.Append(GroupName):
		if err := f("name", value); err != nil {
			return nil, err
		} else {
			g.Name = value
			g.modified = types.TimestampNow()

			list = append(list, kv{GroupName, g.Name})

			g.log(dbc, uid, "update", "name", g.Name, value, "Updated name from %v to %v", original.Name, g.Name)
		}

	case schema.OID(g.OID.Append(GroupDoors)).Contains(oid):
		if m := regexp.MustCompile(`^(?:.*?)\.([0-9]+)$`).FindStringSubmatch(string(oid)); m != nil && len(m) > 1 {
			did := m[1]
			k := schema.DoorsOID.AppendS(did)
			door := catalog.GetV(k, DoorName)

			if err := f(door.(string), value); err != nil {
				return nil, err
			} else {
				if value == "true" {
					g.log(dbc, uid, "update", "door", "", "", "Granted access to %v", door)
				} else {
					g.log(dbc, uid, "update", "door", "", "", "Revoked access to %v", door)
				}

				g.Doors[k] = value == "true"
				g.modified = types.TimestampNow()

				list = append(list, kv{GroupDoors.Append(did), g.Doors[k]})
			}
		}
	}

	if dbc != nil {
		dbc.Updated(g.OID, "", g.Doors)
	}

	list = append(list, kv{GroupStatus, g.Status()})

	return g.toObjects(list, a), nil
}

func (g *Group) delete(a *auth.Authorizator, dbc db.DBC) ([]schema.Object, error) {
	list := []kv{}

	if g != nil {
		if a != nil {
			if err := a.CanDelete(g, auth.Groups); err != nil {
				return nil, err
			}
		}

		g.log(dbc, auth.UID(a), "delete", "group", g.Name, "", "Deleted group %v", g.Name)
		g.deleted = types.TimestampNow()
		g.modified = types.TimestampNow()

		list = append(list, kv{GroupStatus, g.Status()})
		list = append(list, kv{GroupDeleted, g.deleted})

		catalog.DeleteT(g.CatalogGroup, g.OID)
	}

	return g.toObjects(list, a), nil
}

func (g *Group) toObjects(list []kv, a *auth.Authorizator) []schema.Object {
	f := func(g *Group, field string, value interface{}) bool {
		if a != nil {
			if err := a.CanView(g, field, value, auth.Groups); err != nil {
				return false
			}
		}

		return true
	}

	objects := []schema.Object{}

	if !g.IsDeleted() && f(g, "OID", g.OID) {
		catalog.Join(&objects, catalog.NewObject(g.OID, ""))
	}

	for _, v := range list {
		field, _ := lookup[v.field]
		if f(g, field, v.value) {
			catalog.Join(&objects, catalog.NewObject2(g.OID, v.field, v.value))
		}
	}

	return objects
}

func (g Group) serialize() ([]byte, error) {
	record := struct {
		OID      schema.OID      `json:"OID"`
		Name     string          `json:"name,omitempty"`
		Doors    []schema.OID    `json:"doors"`
		Created  types.Timestamp `json:"created,omitempty"`
		Modified types.Timestamp `json:"modified,omitempty"`
	}{
		OID:      g.OID,
		Name:     g.Name,
		Doors:    []schema.OID{},
		Created:  g.created.UTC(),
		Modified: g.modified.UTC(),
	}

	doors := catalog.GetDoors()

	for _, d := range doors {
		if g.Doors[d] {
			record.Doors = append(record.Doors, d)
		}
	}

	return json.Marshal(record)
}

func (g *Group) deserialize(bytes []byte) error {
	created = created.Add(1 * time.Minute)

	record := struct {
		OID      string          `json:"OID"`
		Name     string          `json:"name,omitempty"`
		Doors    []schema.OID    `json:"doors"`
		Created  types.Timestamp `json:"created,omitempty"`
		Modified types.Timestamp `json:"modified,omitempty"`
	}{
		Created: created,
	}

	if err := json.Unmarshal(bytes, &record); err != nil {
		return err
	}

	g.OID = schema.OID(record.OID)
	g.Name = record.Name
	g.Doors = map[schema.OID]bool{}
	g.created = record.Created
	g.modified = record.Modified

	for _, d := range record.Doors {
		g.Doors[schema.OID(d)] = true
	}

	return nil
}

func (g Group) clone() Group {
	group := Group{
		CatalogGroup: catalog.CatalogGroup{
			OID: g.OID,
		},
		Name:     g.Name,
		Doors:    map[schema.OID]bool{},
		created:  g.created,
		modified: g.modified,
		deleted:  g.deleted,
	}

	for k, v := range g.Doors {
		group.Doors[k] = v
	}

	return group
}

func (g *Group) log(dbc db.DBC, uid, op string, field string, before, after any, format string, fields ...any) {
	if dbc != nil {
		dbc.Log(uid, op, g.OID, "group", "", g.Name, field, before, after, format, fields...)
	}
}
