package events

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-lib/uhppoted"
)

type Events struct {
	Events map[uint32]map[uint32]Event `json:"events"`

	file string `json:"-"`
}

var guard sync.RWMutex

func NewEvents() Events {
	return Events{
		Events: map[uint32]map[uint32]Event{},
	}
}

func (ee *Events) Load(file string) error {
	blob := struct {
		Events []json.RawMessage `json:"events"`
	}{
		Events: []json.RawMessage{},
	}

	bytes, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, &blob)
	if err != nil {
		return err
	}

	for _, v := range blob.Events {
		var e Event
		if err := e.deserialize(v); err == nil {
			l, ok := ee.Events[e.DeviceID]
			if !ok {
				l = map[uint32]Event{}
				ee.Events[e.DeviceID] = l
			}

			if _, ok := l[e.Index]; ok {
				return fmt.Errorf("event '%v:%v:%v': duplicate OID (%v)", e.DeviceID, e.Index, e.Timestamp, e.OID)
			}

			l[e.Index] = e
		}
	}

	for _, l := range ee.Events {
		for _, e := range l {
			catalog.PutEvent(e.OID)
		}
	}

	ee.file = file

	return nil
}

func (ee Events) Save() error {
	if err := validate(ee); err != nil {
		return err
	}

	if err := scrub(ee); err != nil {
		return err
	}

	if ee.file == "" {
		return nil
	}

	serializable := struct {
		Events []json.RawMessage `json:"events"`
	}{
		Events: []json.RawMessage{},
	}

	for _, l := range ee.Events {
		for _, e := range l {
			if e.IsValid() && !e.IsDeleted() {
				if record, err := e.serialize(); err == nil && record != nil {
					serializable.Events = append(serializable.Events, record)
				}
			}
		}
	}

	b, err := json.MarshalIndent(serializable, "", "  ")
	if err != nil {
		return err
	}

	tmp, err := os.CreateTemp("", "uhppoted-events.*")
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

	if err := os.MkdirAll(filepath.Dir(ee.file), 0770); err != nil {
		return err
	}

	return os.Rename(tmp.Name(), ee.file)
}

func (ee *Events) Stash() {
}

func (ee Events) Print() {
	if b, err := json.MarshalIndent(ee.Events, "", "  "); err == nil {
		fmt.Printf("----------------- EVENTS\n%s\n", string(b))
	}
}

func (ee *Events) Clone() *Events {
	shadow := Events{
		Events: map[uint32]map[uint32]Event{},
		file:   ee.file,
	}

	for k, v := range ee.Events {
		shadow.Events[k] = map[uint32]Event{}
		for id, e := range v {
			shadow.Events[k][id] = e.clone()
		}
	}

	return &shadow
}

func (ee *Events) AsObjects() []interface{} {
	guard.RLock()
	defer guard.RUnlock()

	objects := []interface{}{}

	for _, v := range ee.Events {
		for _, e := range v {
			if e.IsValid() || e.IsDeleted() {
				if l := e.AsObjects(); l != nil {
					objects = append(objects, l...)
				}
			}
		}
	}

	return objects
}

func (ee *Events) UpdateByOID(auth auth.OpAuth, oid catalog.OID, value string) ([]interface{}, error) {
	if ee == nil {
		return nil, nil
	}

	for k, v := range ee.Events {
		for id, e := range v {
			if e.OID.Contains(oid) {
				objects, err := e.set(auth, oid, value)
				if err == nil {
					ee.Events[k][id] = e
				}

				return objects, err
			}
		}
	}

	return []interface{}{}, nil
}

func (ee *Events) Validate() error {
	return nil
}

func (ee *Events) Received(deviceID uint32, recent []uhppoted.Event, lookup func(uhppoted.Event) (string, string, string)) {
	list, ok := ee.Events[deviceID]
	if !ok {
		list = map[uint32]Event{}
	}

	for _, e := range recent {
		var oid catalog.OID

		if v, ok := list[e.Index]; ok {
			oid = v.OID
		} else {
			oid = catalog.NewEvent()
		}

		list[e.Index] = NewEvent(oid, e, lookup)
	}

	guard.Lock()
	defer guard.Unlock()

	ee.Events[deviceID] = list
}

func validate(ee Events) error {
	return nil
}

func scrub(ee Events) error {
	return nil
}
