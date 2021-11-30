package groups

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/uhppoted/uhppoted-httpd/audit"
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/db"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type Group struct {
	OID   catalog.OID          `json:"OID"`
	Name  string               `json:"name"`
	Doors map[catalog.OID]bool `json:"doors"`

	created types.DateTime
	deleted *types.DateTime
}

const BLANK = "'blank'"

var created = time.Now()

func (g Group) String() string {
	return fmt.Sprintf("%v", g.Name)
}

func (g Group) IsValid() bool {
	if strings.TrimSpace(g.Name) != "" {
		return true
	}

	return false
}

func (g Group) IsDeleted() bool {
	if g.deleted != nil {
		return true
	}

	return false
}

func (g *Group) AsObjects() []interface{} {
	created := g.created.Format("2006-01-02 15:04:05")
	name := g.Name

	objects := []interface{}{
		catalog.NewObject(g.OID, ""),
		catalog.NewObject2(g.OID, GroupStatus, g.status()),
		catalog.NewObject2(g.OID, GroupCreated, created),
		catalog.NewObject2(g.OID, GroupDeleted, g.deleted),
		catalog.NewObject2(g.OID, GroupName, name),
	}

	doors := catalog.GetDoors()
	re := regexp.MustCompile(`^(.*?)(\.[0-9]+)$`)

	for _, door := range doors {
		d := fmt.Sprintf("%v", door)

		if m := re.FindStringSubmatch(d); m != nil && len(m) > 2 {
			did := m[2]
			allowed := g.Doors[door]

			objects = append(objects, catalog.NewObject2(g.OID, GroupDoors.Append(did), allowed))
			objects = append(objects, catalog.NewObject2(g.OID, GroupDoors.Append(did+".1"), door))
		}
	}

	return objects
}

func (g *Group) AsRuleEntity() interface{} {
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

	return &entity
}

func (g *Group) status() types.Status {
	if g.deleted != nil {
		return types.StatusDeleted
	}

	return types.StatusOk
}

func (g *Group) set(auth auth.OpAuth, oid catalog.OID, value string, dbc db.DBC) ([]catalog.Object, error) {
	objects := []catalog.Object{}

	f := func(field string, value interface{}) error {
		if auth != nil {
			return auth.CanUpdateGroup(g, field, value)
		}

		return nil
	}

	if g != nil {
		name := g.Name
		switch {
		case oid == g.OID.Append(GroupName):
			if err := f("name", value); err != nil {
				return nil, err
			} else {
				g.log(auth, "update", g.OID, "name", fmt.Sprintf("Updated name from %v to %v", stringify(g.Name, BLANK), stringify(value, BLANK)), dbc)
				g.Name = value
				objects = append(objects, catalog.NewObject2(g.OID, GroupName, g.Name))
			}

		case catalog.OID(g.OID.Append(GroupDoors)).Contains(oid):
			if m := regexp.MustCompile(`^(?:.*?)\.([0-9]+)$`).FindStringSubmatch(string(oid)); m != nil && len(m) > 1 {
				did := m[1]
				k := catalog.DoorsOID.AppendS(did)
				door := catalog.GetV(k, DoorName)

				if err := f(door.(string), value); err != nil {
					return nil, err
				} else {
					if value == "true" {
						g.log(auth, "update", g.OID, "door", fmt.Sprintf("Granted access to %v", door), dbc)
					} else {
						g.log(auth, "update", g.OID, "door", fmt.Sprintf("Revoked access to %v", door), dbc)
					}

					g.Doors[k] = value == "true"
					objects = append(objects, catalog.NewObject2(g.OID, GroupDoors.Append(did), g.Doors[k]))
				}
			}
		}

		if !g.IsValid() {
			if auth != nil {
				if err := auth.CanDeleteGroup(g); err != nil {
					return nil, err
				}
			}

			g.log(auth, "delete", g.OID, "group", fmt.Sprintf("Deleted group %v", name), dbc)
			now := types.DateTime(time.Now())
			g.deleted = &now
			objects = append(objects, catalog.NewObject(g.OID, "deleted"))

			catalog.Delete(g.OID)
		}

		objects = append(objects, catalog.NewObject2(g.OID, GroupStatus, g.status()))
		objects = append(objects, catalog.NewObject2(g.OID, GroupDeleted, g.deleted))
	}

	return objects, nil
}

func (g Group) serialize() ([]byte, error) {
	record := struct {
		OID     catalog.OID   `json:"OID"`
		Name    string        `json:"name,omitempty"`
		Doors   []catalog.OID `json:"doors"`
		Created string        `json:"created"`
	}{
		OID:     g.OID,
		Name:    g.Name,
		Doors:   []catalog.OID{},
		Created: g.created.Format("2006-01-02 15:04:05"),
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
	datetime := types.DateTime(created)

	record := struct {
		OID     string          `json:"OID"`
		Name    string          `json:"name,omitempty"`
		Doors   []catalog.OID   `json:"doors"`
		Created *types.DateTime `json:"created,omitempty"`
	}{
		Created: &datetime,
	}

	if err := json.Unmarshal(bytes, &record); err != nil {
		return err
	}

	g.OID = catalog.OID(record.OID)
	g.Name = record.Name
	g.Doors = map[catalog.OID]bool{}
	g.created = *record.Created

	for _, d := range record.Doors {
		g.Doors[catalog.OID(d)] = true
	}

	return nil
}

func (g Group) clone() Group {
	group := Group{
		OID:     g.OID,
		Name:    g.Name,
		Doors:   map[catalog.OID]bool{},
		created: g.created,
		deleted: g.deleted,
	}

	for k, v := range g.Doors {
		group.Doors[k] = v
	}

	return group
}

func (g *Group) log(auth auth.OpAuth, operation string, OID catalog.OID, field string, description string, dbc db.DBC) {
	uid := ""
	if auth != nil {
		uid = auth.UID()
	}

	record := audit.AuditRecord{
		UID:       uid,
		OID:       OID,
		Component: "group",
		Operation: operation,
		Details: audit.Details{
			ID:          "",
			Name:        stringify(g.Name, ""),
			Field:       field,
			Description: description,
		},
	}

	if dbc != nil {
		dbc.Write(record)
	}
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
		if i != nil {
			s = fmt.Sprintf("%v", i)
		}
	}

	if s != "" {
		return s
	}

	return defval
}
