package groups

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/uhppoted/uhppoted-httpd/audit"
	"github.com/uhppoted/uhppoted-httpd/auth"
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

func (gg Groups) Save() error {
	if err := validate(gg); err != nil {
		return err
	}

	if err := scrub(gg); err != nil {
		return err
	}

	if gg.file == "" {
		return nil
	}

	serializable := struct {
		Groups []json.RawMessage `json:"groups"`
	}{
		Groups: []json.RawMessage{},
	}

	for _, g := range gg.Groups {
		if g.IsValid() && !g.IsDeleted() {
			if record, err := g.serialize(); err == nil && record != nil {
				serializable.Groups = append(serializable.Groups, record)
			}
		}
	}

	b, err := json.MarshalIndent(serializable, "", "  ")
	if err != nil {
		return err
	}

	tmp, err := os.CreateTemp("", "uhppoted-groups.*")
	if err != nil {
		return err
	}

	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(b); err != nil {
		return err
	}

	if err := tmp.Close(); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(gg.file), 0770); err != nil {
		return err
	}

	return os.Rename(tmp.Name(), gg.file)
}

func (gg Groups) Print() {
	if b, err := json.MarshalIndent(gg.Groups, "", "  "); err == nil {
		fmt.Printf("----------------- GROUPS\n%s\n", string(b))
	}
}

func (gg *Groups) Clone() *Groups {
	shadow := Groups{
		Groups: map[catalog.OID]Group{},
		file:   gg.file,
	}

	for k, v := range gg.Groups {
		shadow.Groups[k] = v.clone()
	}

	return &shadow
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

func (gg *Groups) UpdateByOID(auth auth.OpAuth, oid string, value string) ([]interface{}, error) {
	if gg == nil {
		return nil, nil
	}

	for k, g := range gg.Groups {
		if g.OID.Contains(oid) {
			objects, err := g.set(auth, oid, value)
			if err == nil {
				gg.Groups[k] = g
			}

			return objects, err
		}
	}

	objects := []interface{}{}

	if oid == "<new>" {
		if g, err := gg.add(auth, Group{}); err != nil {
			return nil, err
		} else if g == nil {
			return nil, fmt.Errorf("Failed to add 'new' group")
		} else {
			g.log(auth, "add", g.OID, "group", "", "")
			gg.Groups[g.OID] = *g
			objects = append(objects, object{
				OID:   stringify(g.OID),
				Value: "new",
			})
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

func (gg *Groups) add(auth auth.OpAuth, g Group) (*Group, error) {
	oid := catalog.NewGroup()

	record := g.clone()
	record.OID = oid
	record.created = time.Now()

	if auth != nil {
		if err := auth.CanAddGroup(&record); err != nil {
			return nil, err
		}
	}

	return &record, nil
}

func validate(gg Groups) error {
	names := map[string]string{}

	for k, g := range gg.Groups {
		if g.deleted != nil {
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
