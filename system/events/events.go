package events

import (
	"encoding/json"
	"fmt"
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
	return nil
}

func (ee Events) Save() error {
	return nil
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

func (ee *Events) Received(deviceID uint32, recent []uhppoted.Event) {
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

		list[e.Index] = NewEvent(oid, e)
	}

	guard.Lock()
	defer guard.Unlock()

	ee.Events[deviceID] = list
}
