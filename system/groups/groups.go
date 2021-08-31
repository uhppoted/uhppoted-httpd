package groups

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/uhppoted/uhppoted-httpd/audit"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
)

type Groups struct {
	Groups map[catalog.OID]Group `json:"groups"`

	file string `json:"-"`
}

type object catalog.Object

var trail audit.Trail

func SetAuditTrail(t audit.Trail) {
	trail = t
}

func NewGroups() Groups {
	return Groups{
		Groups: map[catalog.OID]Group{},
	}
}

func (gg *Groups) Load(file string) error {
	blob := struct {
		Groups []json.RawMessage `json:"groups"`
	}{
		Groups: []json.RawMessage{},
	}

	bytes, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, &blob)
	if err != nil {
		return err
	}

	for _, v := range blob.Groups {
		var g Group
		if err := g.deserialize(v); err == nil {
			if _, ok := gg.Groups[g.OID]; ok {
				return fmt.Errorf("group '%v': duplicate OID (%v)", g.Name, g.OID)
			}

			gg.Groups[g.OID] = g
		}
	}

	for _, g := range gg.Groups {
		catalog.PutGroup(g.OID)
		catalog.PutV(g.OID.Append(GroupName), g.Name, false)
		catalog.PutV(g.OID.Append(GroupCreated), g.created, false)
	}

	gg.file = file

	return nil
}

func (gg Groups) Print() {
	if b, err := json.MarshalIndent(gg.Groups, "", "  "); err == nil {
		fmt.Printf("----------------- GROUPS\n%s\n", string(b))
	}
}

func (gg *Groups) AsObjects() []interface{} {
	objects := []interface{}{}

	for _, g := range gg.Groups {
		if g.IsValid() || g.IsDeleted() {
			if l := g.AsObjects(); l != nil {
				objects = append(objects, l...)
			}
		}
	}

	return objects
}
