package groups

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/uhppoted/uhppoted-httpd/audit"
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type Group struct {
	OID  catalog.OID `json:"OID"`
	Name string      `json:"name"`

	created time.Time
	deleted *time.Time
}

const GroupName = catalog.GroupName
const GroupCreated = catalog.GroupCreated

var created = time.Now()

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

func (g Group) clone() Group {
	return Group{
		OID:     g.OID,
		Name:    g.Name,
		created: g.created,
		deleted: g.deleted,
	}
}

func (g *Group) AsObjects() []interface{} {
	created := g.created.Format("2006-01-02 15:04:05")
	status := stringify(types.StatusOk)
	name := stringify(g.Name)

	if g.deleted != nil {
		status = stringify(types.StatusDeleted)
	}

	objects := []interface{}{
		object{OID: string(g.OID), Value: status},
		object{OID: g.OID.Append(GroupCreated), Value: created},
		object{OID: g.OID.Append(GroupName), Value: name},
	}

	return objects
}

func (g *Group) set(auth auth.OpAuth, oid string, value string) ([]interface{}, error) {
	objects := []interface{}{}

	f := func(field string, value interface{}) error {
		// if auth == nil {
		// 	return nil
		// }
		//
		// return auth.CanUpdateGroup(d, field, value)

		return nil
	}

	if g != nil {
		// name := stringify(g.Name)

		switch oid {
		case g.OID.Append(GroupName):
			if err := f("name", value); err != nil {
				return nil, err
			} else {
				g.log(auth, "update", g.OID, "name", stringify(g.Name), value)
				g.Name = value
				objects = append(objects, object{
					OID:   g.OID.Append(GroupName),
					Value: stringify(g.Name),
				})
			}
		}

		// if !g.IsValid() {
		// 	if auth != nil {
		// 		if err := auth.CanDeleteGroup(d); err != nil {
		// 			return nil, err
		// 		}
		// 	}
		//
		// 	g.log(auth, "delete", g.OID, "name", name, "")
		// 	now := time.Now()
		// 	g.deleted = &now
		//
		// 	objects = append(objects, object{
		// 		OID:   g.OID,
		// 		Value: "deleted",
		// 	})
		//
		// 	catalog.Delete(g.OID)
		// }
	}

	return objects, nil
}

func (g Group) serialize() ([]byte, error) {
	record := struct {
		OID     catalog.OID `json:"OID"`
		Name    string      `json:"name,omitempty"`
		Created string      `json:"created"`
	}{
		OID:     g.OID,
		Name:    g.Name,
		Created: g.created.Format("2006-01-02 15:04:05"),
	}

	return json.Marshal(record)
}

func (g *Group) deserialize(bytes []byte) error {
	created = created.Add(1 * time.Minute)

	record := struct {
		OID     string `json:"OID"`
		Name    string `json:"name,omitempty"`
		Created string `json:"created"`
	}{}

	if err := json.Unmarshal(bytes, &record); err != nil {
		return err
	}

	g.OID = catalog.OID(record.OID)
	g.Name = record.Name
	g.created = created

	if t, err := time.Parse("2006-01-02 15:04:05", record.Created); err == nil {
		g.created = t
	}

	return nil
}

func (g *Group) log(auth auth.OpAuth, operation string, OID catalog.OID, field, current, value string) {
	type info struct {
		OID     string `json:"OID"`
		Group   string `json:"group"`
		Field   string `json:"field"`
		Current string `json:"current"`
		Updated string `json:"new"`
	}

	uid := ""
	if auth != nil {
		uid = auth.UID()
	}

	if trail != nil {
		record := audit.LogEntry{
			UID:       uid,
			Module:    stringify(OID),
			Operation: operation,
			Info: info{
				OID:     stringify(OID),
				Group:   stringify(g.Name),
				Field:   field,
				Current: current,
				Updated: value,
			},
		}

		trail.Write(record)
	}
}

func stringify(i interface{}) string {
	switch v := i.(type) {
	case *uint32:
		if v != nil {
			return fmt.Sprintf("%v", *v)
		}

	case *string:
		if v != nil {
			return fmt.Sprintf("%v", *v)
		}

	default:
		if i != nil {
			return fmt.Sprintf("%v", i)
		}
	}

	return ""
}
