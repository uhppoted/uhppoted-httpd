package interfaces

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	lib "github.com/uhppoted/uhppote-core/types"
	"github.com/uhppoted/uhppoted-lib/acl"

	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/log"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	"github.com/uhppoted/uhppoted-httpd/system/db"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type Interfaces struct {
	lans map[schema.OID]*LAN
	ch   chan types.EventsList
}

var guards = sync.Map{}
var guard sync.RWMutex

func NewInterfaces(ch chan types.EventsList) Interfaces {
	return Interfaces{
		lans: map[schema.OID]*LAN{},
		ch:   ch,
	}
}

func (ii *Interfaces) AsObjects(a *auth.Authorizator) []schema.Object {
	objects := []schema.Object{}

	for _, l := range ii.lans {
		if l.IsValid() {
			catalog.Join(&objects, l.AsObjects(a)...)
		}
	}

	return objects
}

func (ii *Interfaces) Create(a *auth.Authorizator, oid schema.OID, value string, dbc db.DBC) ([]schema.Object, error) {
	return []schema.Object{}, nil
}

func (ii *Interfaces) Update(a *auth.Authorizator, oid schema.OID, value string, dbc db.DBC) ([]schema.Object, error) {
	objects := []schema.Object{}
	if ii != nil {

		for _, l := range ii.lans {
			if l != nil && l.OID.Contains(oid) {
				return l.set(a, oid, value, dbc)
			}
		}

	}

	return objects, nil
}

func (ii *Interfaces) Delete(a *auth.Authorizator, oid schema.OID, dbc db.DBC) ([]schema.Object, error) {
	return []schema.Object{}, nil
}

func (ii *Interfaces) LAN() (LAN, bool) {
	for _, v := range ii.lans {
		if v != nil {
			return *v, true
		}
	}

	return LAN{}, false
}

func (ii *Interfaces) Load(blob json.RawMessage) error {
	rs := []json.RawMessage{}
	if err := json.Unmarshal(blob, &rs); err != nil {
		return err
	}

	for _, v := range rs {
		var l LAN
		if err := l.deserialize(v); err == nil {
			if _, ok := ii.lans[l.OID]; ok {
				return fmt.Errorf("card '%v': duplicate OID (%v)", l.Name, l.OID)
			}

			l.ch = ii.ch
			ii.lans[l.OID] = &l
		}
	}

	for _, i := range ii.lans {
		catalog.PutT(i.CatalogInterface)
	}

	return nil
}

func (ii Interfaces) Save() (json.RawMessage, error) {
	if err := ii.Validate(); err != nil {
		return nil, err
	}

	serializable := []json.RawMessage{}
	for _, l := range ii.lans {
		if l.IsValid() && !l.IsDeleted() {
			if record, err := l.serialize(); err == nil && record != nil {
				serializable = append(serializable, record)
			}
		}
	}

	return json.MarshalIndent(serializable, "", "  ")
}

func (ii Interfaces) Print() {
	serializable := []json.RawMessage{}
	for _, l := range ii.lans {
		if l.IsValid() && !l.IsDeleted() {
			if record, err := l.serialize(); err == nil && record != nil {
				serializable = append(serializable, record)
			}
		}
	}

	if b, err := json.MarshalIndent(serializable, "", "  "); err == nil {
		fmt.Printf("----------------- INTERFACES\n%s\n", string(b))
	}
}

func (ii *Interfaces) Clone() Interfaces {
	guard.RLock()
	defer guard.RUnlock()

	shadow := Interfaces{
		lans: map[schema.OID]*LAN{},
	}

	for k, v := range ii.lans {
		clone := v.Clone()
		shadow.lans[k] = &clone
	}

	return shadow
}

func (ii Interfaces) Validate() error {
	names := map[string]string{}

	for k, l := range ii.lans {
		if l.IsDeleted() {
			continue
		}

		if l.OID == "" {
			return fmt.Errorf("Invalid LAN OID (%v)", l.OID)
		}

		if k != l.OID {
			return fmt.Errorf("LAN %s: mismatched LAN OID %v (expected %v)", l.Name, l.OID, k)
		}

		if err := l.validate(); err != nil {
			return err
		}

		n := strings.TrimSpace(strings.ToLower(l.Name))
		if v, ok := names[n]; ok && n != "" {
			return fmt.Errorf("'%v': duplicate LAN name (%v)", l.Name, v)
		}

		names[n] = l.Name
	}

	return nil
}

func (ii *Interfaces) Search(controllers []types.IController) []uint32 {
	var mutex sync.Mutex
	var wg sync.WaitGroup
	var found = map[uint32]struct{}{}

	f := func(lan *LAN) {
		defer wg.Done()

		if list, err := lan.search(controllers); err != nil {
			log.Warnf("%v", err)
		} else {
			mutex.Lock()
			defer mutex.Unlock()
			for _, v := range list {
				found[v] = struct{}{}
			}
		}

	}

	for _, l := range ii.lans {
		lan := l
		wg.Add(1)
		go f(lan)
	}

	wg.Wait()

	list := []uint32{}
	for k, _ := range found {
		list = append(list, k)
	}

	return list
}

func (ii *Interfaces) Refresh(controllers []types.IController) {
	if lan, ok := ii.LAN(); ok {
		var wg sync.WaitGroup

		for _, c := range controllers {
			wg.Add(1)

			controller := c

			go func(v types.IController) {
				defer wg.Done()
				lan.refresh(controller)
			}(controller)
		}

		wg.Wait()
	}
}

func (ii *Interfaces) GetEvents(controllers []types.IController, missing map[uint32][]types.Interval) {
	if lan, ok := ii.LAN(); ok {
		var wg sync.WaitGroup

		for _, c := range controllers {
			wg.Add(1)

			controller := c
			intervals := missing[c.ID()]

			go func(v types.IController) {
				defer wg.Done()
				lan.getEvents(controller, intervals)
			}(controller)
		}

		wg.Wait()
	}
}

func (ii *Interfaces) SetTime(controller types.IController, t time.Time) {
	if lan, ok := ii.LAN(); ok {
		lan.setTime(controller, t)
	}
}

func (ii *Interfaces) SetDoor(controller types.IController, door uint8, mode lib.ControlState, delay uint8) {
	if lan, ok := ii.LAN(); ok {
		if err := lan.setDoor(controller, door, mode, delay); err != nil {
			log.Warnf("%v", err)
		}
	}
}

func (ii *Interfaces) SetDoorControl(controller types.IController, door uint8, mode lib.ControlState) {
	if lan, ok := ii.LAN(); ok {
		if err := lan.setDoor(controller, door, mode, 0); err != nil {
			log.Warnf("%v", err)
		} else if oid, ok := controller.Door(door); ok {
			catalog.PutV(oid, DoorControl, mode)
			catalog.PutV(oid, DoorControlModified, false)
		}
	}
}

func (ii *Interfaces) SetDoorDelay(controller types.IController, door uint8, delay uint8) {
	if lan, ok := ii.LAN(); ok {
		if err := lan.setDoor(controller, door, lib.ModeUnknown, delay); err != nil {
			log.Warnf("%v", err)
		} else if oid, ok := controller.Door(door); ok {
			catalog.PutV(oid, DoorDelay, delay)
			catalog.PutV(oid, DoorDelayModified, false)
		}
	}
}

func (ii *Interfaces) PutCard(controller types.IController, card uint32, from, to lib.Date, permissions map[uint8]uint8) {
	if lan, ok := ii.LAN(); ok {
		lan.putCard(controller, card, from, to, permissions)
	}
}

func (ii *Interfaces) DeleteCard(controller types.IController, card uint32) {
	if lan, ok := ii.LAN(); ok {
		lan.deleteCard(controller, card)
	}
}

func (ii *Interfaces) CompareACL(controllers []types.IController, permissions acl.ACL) (map[uint32]acl.Diff, error) {
	if lan, ok := ii.LAN(); ok {
		return lan.compareACL(controllers, permissions)
	}

	return nil, nil
}

func (ii *Interfaces) add(auth auth.OpAuth, l LAN) (*LAN, error) {
	return nil, fmt.Errorf("NOT SUPPORTED")
}

func lock(id uint32) {
	g := sync.Mutex{}
	if guard, _ := guards.LoadOrStore(id, &g); guard != nil {
		guard.(*sync.Mutex).Lock()
	}
}

func unlock(id uint32) {
	if guard, ok := guards.Load(id); ok && guard != nil {
		guard.(*sync.Mutex).Unlock()
	}
}
