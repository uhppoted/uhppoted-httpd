package doors

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/uhppoted/uhppoted-httpd/audit"
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
)

type Doors struct {
	Doors map[string]Door `json:"doors"`

	file string `json:"-"`
}

type object catalog.Object

var guard sync.Mutex

var trail audit.Trail

func SetAuditTrail(t audit.Trail) {
	trail = t
}

func NewDoors() Doors {
	return Doors{
		Doors: map[string]Door{},
	}
}

func (dd *Doors) AsObjects() []interface{} {
	objects := []interface{}{}

	for _, d := range dd.Doors {
		if d.IsValid() || d.IsDeleted() {
			if l := d.AsObjects(); l != nil {
				objects = append(objects, l...)
			}
		}
	}

	return objects
}

func (dd *Doors) Load(file string) error {
	blob := struct {
		Doors []json.RawMessage `json:"doors"`
	}{
		Doors: []json.RawMessage{},
	}

	bytes, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, &blob)
	if err != nil {
		return err
	}

	for _, v := range blob.Doors {
		var d Door
		if err := d.deserialize(v); err == nil {
			if _, ok := dd.Doors[d.OID]; ok {
				return fmt.Errorf("door '%v': duplicate OID (%v)", d.Name, d.OID)
			}

			dd.Doors[d.OID] = d
		}
	}

	for _, v := range dd.Doors {
		catalog.PutDoor(v.OID)
		catalog.PutV(v.OID+DoorDelayConfigured, v.delay, false)
		catalog.PutV(v.OID+DoorControlConfigured, v.mode, false)
	}

	dd.file = file

	return nil
}

func (dd Doors) Save() error {
	if err := validate(dd); err != nil {
		return err
	}

	if err := scrub(dd); err != nil {
		return err
	}

	if dd.file == "" {
		return nil
	}

	serializable := struct {
		Doors []json.RawMessage `json:"doors"`
	}{
		Doors: []json.RawMessage{},
	}

	for _, d := range dd.Doors {
		if d.IsValid() && !d.IsDeleted() {
			if record, err := d.serialize(); err == nil && record != nil {
				serializable.Doors = append(serializable.Doors, record)
			}
		}
	}

	b, err := json.MarshalIndent(serializable, "", "  ")
	if err != nil {
		return err
	}

	tmp, err := os.CreateTemp("", "uhppoted-doors.*")
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

	if err := os.MkdirAll(filepath.Dir(dd.file), 0770); err != nil {
		return err
	}

	return os.Rename(tmp.Name(), dd.file)
}

func (dd *Doors) Sweep(retention time.Duration) {
	if dd != nil {
		cutoff := time.Now().Add(-retention)
		for i, v := range dd.Doors {
			if v.deleted != nil && v.deleted.Before(cutoff) {
				delete(dd.Doors, i)
			}
		}
	}
}

func (dd Doors) Find(name string) (Door, bool) {
	clean := func(s string) string {
		return strings.ToLower(regexp.MustCompile(`\s+`).ReplaceAllString(s, ""))
	}

	p := clean(name)

	if p != "" {
		for _, d := range dd.Doors {
			if p == clean(d.Name) {
				return d, true
			}
		}
	}

	return Door{}, false
}

func (dd Doors) Print() {
	if b, err := json.MarshalIndent(dd.Doors, "", "  "); err == nil {
		fmt.Printf("----------------- DOORS\n%s\n", string(b))
	}
}

func (dd *Doors) UpdateByOID(auth auth.OpAuth, oid string, value string) ([]interface{}, error) {
	if dd == nil {
		return nil, nil
	}

	for k, d := range dd.Doors {
		if strings.HasPrefix(oid, d.OID+".") {
			objects, err := d.set(auth, oid, value)
			if err == nil {
				dd.Doors[k] = d
			}

			return objects, err
		}
	}

	objects := []interface{}{}

	if oid == "<new>" {
		if d, err := dd.add(auth, Door{}); err != nil {
			return nil, err
		} else if d == nil {
			return nil, fmt.Errorf("Failed to add 'new' door")
		} else {
			d.log(auth, "add", d.OID, "door", "", "")
			dd.Doors[d.OID] = *d
			objects = append(objects, object{
				OID:   d.OID,
				Value: "new",
			})
		}
	}

	return objects, nil
}

func (dd *Doors) add(auth auth.OpAuth, d Door) (*Door, error) {
	oid := catalog.NewDoor()

	record := d.clone()
	record.OID = oid
	record.created = time.Now()

	if auth != nil {
		if err := auth.CanAddDoor(&record); err != nil {
			return nil, err
		}
	}

	return &record, nil
}

func (dd *Doors) Clone() *Doors {
	shadow := Doors{
		Doors: map[string]Door{},
		file:  dd.file,
	}

	for k, v := range dd.Doors {
		shadow.Doors[k] = v.clone()
	}

	return &shadow
}

func (dd *Doors) Validate() error {
	if dd != nil {
		return validate(*dd)
	}

	return nil
}

func validate(dd Doors) error {
	names := map[string]string{}

	for k, d := range dd.Doors {
		if d.deleted != nil {
			continue
		}

		if d.OID == "" {
			return fmt.Errorf("Invalid door OID (%v)", d.OID)
		}

		if k != d.OID {
			return fmt.Errorf("Door %s: mismatched door OID %v (expected %v)", d.Name, d.OID, k)
		}

		n := strings.TrimSpace(strings.ToLower(d.Name))
		if v, ok := names[n]; ok && n != "" {
			return fmt.Errorf("'%v': duplicate door name (%v)", d.Name, v)
		}

		names[n] = d.Name
	}

	return nil
}

func scrub(dd Doors) error {
	return nil
}
