package interfaces

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

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

type Controller interface {
	OIDx() schema.OID
	Name() string
	ID() uint32
	EndPoint() *net.UDPAddr
	TimeZone() *time.Location
	Door(uint8) (schema.OID, bool)
}

type Events interface {
	Indices(deviceID uint32) (first uint32, last uint32)
}

const BLANK = "'blank'"

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

func (ii *Interfaces) UpdateByOID(a *auth.Authorizator, oid schema.OID, value string, dbc db.DBC) ([]schema.Object, error) {
	if ii == nil {
		return nil, nil
	}

	for _, l := range ii.lans {
		if l != nil && l.OID.Contains(oid) {
			return l.set(a, oid, value, dbc)
		}
	}

	objects := []schema.Object{}

	if oid == "<new>" {
		if l, err := ii.add(a, LAN{}); err != nil {
			return nil, err
		} else if l == nil {
			return nil, fmt.Errorf("Failed to add 'new' interface")
		} else {
			l.log(auth.UID(a), "add", l.OID, "interface", fmt.Sprintf("Added 'new' interface"), "", "", dbc)
			catalog.Join(&objects, catalog.NewObject(l.OID, "new"))
			catalog.Join(&objects, catalog.NewObject2(l.OID, LANStatus, "new"))
			catalog.Join(&objects, catalog.NewObject2(l.OID, LANCreated, l.created))
		}
	}

	return objects, nil
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

	for _, v := range ii.lans {
		catalog.PutT(v.CatalogInterface, v.OID)
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

func (ii *Interfaces) Search(controllers []Controller) []uint32 {
	var mutex sync.Mutex
	var wg sync.WaitGroup
	var found = map[uint32]struct{}{}

	f := func(lan *LAN) {
		defer wg.Done()

		if list, err := lan.Search(controllers); err != nil {
			log.Warnf("%v", err)
		} else {
			mutex.Lock()
			defer mutex.Unlock()
			for _, v := range list {
				found[v] = struct{}{}
			}
		}

	}

	for _, lan := range ii.lans {
		wg.Add(1)
		go f(lan) // NTS: lan is a pointer so it's more or less ok to reuse the loop variable
	}

	wg.Wait()

	list := []uint32{}
	for k, _ := range found {
		list = append(list, k)
	}

	return list
}

func (ii *Interfaces) Refresh(controllers []Controller) {
	if lan, ok := ii.LAN(); ok {
		var wg sync.WaitGroup

		for _, c := range controllers {
			wg.Add(1)

			controller := c

			go func(v Controller) {
				defer wg.Done()
				lan.Refresh(controller)
			}(controller)
		}

		wg.Wait()
	}
}

func (ii *Interfaces) GetEvents(controllers []Controller, missing map[uint32][]types.Interval) {
	if lan, ok := ii.LAN(); ok {
		var wg sync.WaitGroup

		for _, c := range controllers {
			wg.Add(1)

			controller := c
			intervals := missing[c.ID()]

			go func(v Controller) {
				defer wg.Done()
				lan.GetEvents(controller, intervals)
			}(controller)
		}

		wg.Wait()
	}
}

func (ii *Interfaces) add(auth auth.OpAuth, l LAN) (*LAN, error) {
	return nil, fmt.Errorf("NOT SUPPORTED")
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
		s = fmt.Sprintf("%v", i)
	}

	if s != "" {
		return s
	}

	return defval
}
