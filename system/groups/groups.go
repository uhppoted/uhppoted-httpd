package groups

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/db"
)

type Groups struct {
	Groups map[catalog.OID]Group `json:"groups"`

	file string `json:"-"`
}

var guard sync.RWMutex

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

	if err := json.Unmarshal(bytes, &blob); err != nil {
		return err
	}

	var index uint32 = 1

	for _, v := range blob.Groups {
		var g Group
		if err := g.deserialize(v); err == nil {
			if _, ok := gg.Groups[g.OID]; ok {
				return fmt.Errorf("group '%v': duplicate OID (%v)", g.Name, g.OID)
			}

			g.Index = index
			gg.Groups[g.OID] = g

			index++
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

	keys := []catalog.OID{}
	for _, g := range gg.Groups {
		keys = append(keys, g.OID)
	}

	sort.SliceStable(keys, func(i, j int) bool { return gg.Groups[keys[i]].Index < gg.Groups[keys[j]].Index })

	for _, k := range keys {
		if g, ok := gg.Groups[k]; ok && g.IsValid() && !g.IsDeleted() {
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

func (gg *Groups) Stash() {
	if gg != nil {
		for _, g := range gg.Groups {
			g.stash()
		}
	}
}

func (gg Groups) Print() {
	if b, err := json.MarshalIndent(gg.Groups, "", "  "); err == nil {
		fmt.Printf("----------------- GROUPS\n%s\n", string(b))
	}
}

func (gg *Groups) Clone() Groups {
	shadow := Groups{
		Groups: map[catalog.OID]Group{},
		file:   gg.file,
	}

	for k, v := range gg.Groups {
		shadow.Groups[k] = v.clone()
	}

	return shadow
}

func (gg *Groups) AsObjects() []interface{} {
	guard.RLock()
	defer guard.RUnlock()

	objects := []interface{}{}

	keys := []catalog.OID{}
	for _, g := range gg.Groups {
		keys = append(keys, g.OID)
	}

	sort.SliceStable(keys, func(i, j int) bool { return gg.Groups[keys[i]].Index < gg.Groups[keys[j]].Index })

	for _, k := range keys {
		if g, ok := gg.Groups[k]; ok {
			if g.IsValid() || g.IsDeleted() {
				if l := g.AsObjects(); l != nil {
					objects = append(objects, l...)
				}
			}
		}
	}

	return objects
}

func (gg *Groups) UpdateByOID(auth auth.OpAuth, oid catalog.OID, value string, dbc db.DBC) ([]interface{}, error) {
	if gg == nil {
		return nil, nil
	}

	for k, g := range gg.Groups {
		if g.OID.Contains(oid) {
			objects, err := g.set(auth, oid, value, dbc)
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
			g.log(auth, "add", g.OID, "group", "Added <new> group", dbc)

			g.Index = uint32(len(gg.Groups) + 1)
			for _, p := range gg.Groups {
				if g.Index < p.Index {
					g.Index = uint32(p.Index + 1)
				}
			}

			gg.Groups[g.OID] = *g
			objects = append(objects, catalog.NewObject(g.OID, "new"))
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
