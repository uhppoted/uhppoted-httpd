package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/db"
	"github.com/uhppoted/uhppoted-httpd/system/doors"
	"github.com/uhppoted/uhppoted-httpd/system/interfaces"
	"github.com/uhppoted/uhppoted-httpd/types"
	"github.com/uhppoted/uhppoted-lib/acl"
	"github.com/uhppoted/uhppoted-lib/config"
	"github.com/uhppoted/uhppoted-lib/uhppoted"
)

type ControllerSet struct {
	Controllers []*Controller `json:"controllers"`
	LAN         *LAN          `json:"LAN"`
}

type Callback interface {
	Append(deviceID uint32, events []uhppoted.Event)
}

const BLANK = "'blank'"

var guard sync.RWMutex

var windows = struct {
	deviceOk        time.Duration
	deviceUncertain time.Duration
	systime         time.Duration
	cacheExpiry     time.Duration
}{
	deviceOk:        60 * time.Second,
	deviceUncertain: 300 * time.Second,
	systime:         300 * time.Second,
	cacheExpiry:     120 * time.Second,
}

func SetWindows(ok, uncertain, systime, cacheExpiry time.Duration) {
	windows.deviceOk = ok
	windows.deviceUncertain = uncertain
	windows.systime = systime
	windows.cacheExpiry = cacheExpiry
}

func NewControllerSet() ControllerSet {
	return ControllerSet{
		Controllers: []*Controller{},
	}
}

func (cc *ControllerSet) Init(interfaces interfaces.Interfaces) {
	for _, v := range interfaces.LANs {
		if v != nil {
			cc.LAN = &LAN{
				*v,
			}

			break
		}
	}
}

func (cc *ControllerSet) Load(blob json.RawMessage) error {
	rs := []json.RawMessage{}
	if err := json.Unmarshal(blob, &rs); err != nil {
		return err
	}

	cc.Controllers = []*Controller{}
	for _, v := range rs {
		var c Controller
		if err := c.deserialize(v); err != nil {
			warn(err)
		} else {
			cc.Controllers = append(cc.Controllers, &c)
		}
	}

	for _, c := range cc.Controllers {
		if c.deviceID != nil && *c.deviceID != 0 {
			oid := c.OID()
			catalog.PutController(*c.deviceID, oid)
			catalog.PutV(oid, ControllerName, c.name)
			catalog.PutV(oid, ControllerDoor1, c.Doors[1])
			catalog.PutV(oid, ControllerDoor2, c.Doors[2])
			catalog.PutV(oid, ControllerDoor3, c.Doors[3])
			catalog.PutV(oid, ControllerDoor4, c.Doors[4])
		}
	}

	return nil
}

func (cc *ControllerSet) Save() (json.RawMessage, error) {
	if cc == nil {
		return nil, nil
	}

	if err := validate(*cc); err != nil {
		return nil, err
	}

	if err := scrub(cc); err != nil {
		return nil, err
	}

	serializable := []json.RawMessage{}
	for _, c := range cc.Controllers {
		if bytes, err := c.serialize(); err == nil && bytes != nil {
			serializable = append(serializable, bytes)
		}
	}

	return json.MarshalIndent(serializable, "", "  ")
}

func (cc *ControllerSet) AsObjects() []interface{} {
	objects := []interface{}{}

	for _, c := range cc.Controllers {
		if c.IsValid() {
			if l := c.AsObjects(); l != nil {
				objects = append(objects, l...)
			}
		}
	}

	return objects
}

func (cc *ControllerSet) Sweep(retention time.Duration) {
	if cc == nil {
		return
	}

	cutoff := time.Now().Add(-retention)
	for i, v := range cc.Controllers {
		if v.deleted != nil && v.deleted.Before(cutoff) {
			cc.Controllers = append(cc.Controllers[:i], cc.Controllers[i+1:]...)
		}
	}
}

func (cc *ControllerSet) Print() {
	if b, err := json.MarshalIndent(cc, "", "  "); err == nil {
		fmt.Printf("----------------- CONTROLLERS\n%s\n", string(b))
	}
}

func (cc *ControllerSet) UpdateByOID(auth auth.OpAuth, oid catalog.OID, value string, dbc db.DBC) ([]catalog.Object, error) {
	if cc == nil {
		return nil, nil
	}

	for _, c := range cc.Controllers {
		if c != nil && c.OID().Contains(oid) {
			return c.set(auth, oid, value, dbc)
		}
	}

	objects := []catalog.Object{}

	if oid == "<new>" {
		if c, err := cc.add(auth, Controller{}); err != nil {
			return nil, err
		} else if c == nil {
			return nil, fmt.Errorf("Failed to add 'new' controller")
		} else {
			OID := c.OID()
			c.log(auth, "add", OID, "controller", fmt.Sprintf("Added 'new' controller"), "", "", dbc)
			objects = append(objects, catalog.NewObject(OID, "new"))
			objects = append(objects, catalog.NewObject2(OID, ControllerStatus, "new"))
			objects = append(objects, catalog.NewObject2(OID, ControllerCreated, c.created))
		}
	}

	return objects, nil
}

func (cc *ControllerSet) Find(deviceID uint32) *Controller {
	if deviceID != 0 {
		for _, c := range cc.Controllers {
			if c.deviceID != nil && *c.deviceID == deviceID {
				return c
			}
		}
	}

	return nil
}

func (cc *ControllerSet) Refresh(callback Callback) {
	cc.LAN.refresh(cc.Controllers, callback)

		// ... add 'found' controllers to list
	//loop:
	//	for k, _ := range cache.cache {
	//		for _, c := range cc.Controllers {
	//			if c.deviceID != nil && *c.deviceID == k && c.deleted == nil {
	//				continue loop
	//			}
	//		}
	//
	//		id := k
	//		oid := catalog.NewController(k)
	//
	//		cc.Controllers = append(cc.Controllers, &Controller{
	//			OID:          catalog.OID(oid),
	//			deviceID:     &id,
	//			created:      types.DateTime(time.Now()),
	//			unconfigured: true,
	//		})
	//	}

    // ... 'expire' controller values that haven't been updated in a while
	expired := time.Now().Add(-windows.cacheExpiry)
	for _, c := range cc.Controllers {
		touched := time.Time(c.created)

		if v := catalog.GetV(c.OID(), ControllerTouched); v != nil {
			if t, ok := v.(time.Time); ok {
				touched = t
			}
		}

		if touched.Before(expired) {
			catalog.PutV(c.OID(), ControllerEndpointAddress, nil)
			catalog.PutV(c.OID(), ControllerDateTimeCurrent, nil)
			catalog.PutV(c.OID(), ControllerCardsCount, nil)
			catalog.PutV(c.OID(), ControllerCardsStatus, types.StatusUnknown)
			catalog.PutV(c.OID(), ControllerEventsCount, nil)

			for _, d := range []uint8{1, 2, 3, 4} {
				if door, ok := c.Door(d); ok {
					catalog.PutV(door, DoorDelay, nil)
					catalog.PutV(door, DoorControl, nil)
				}
			}


			log.Printf("Controller %v cached values expired", c)
		}
	}
}

func (cc *ControllerSet) Clone() ControllerSet {
	guard.RLock()
	defer guard.RUnlock()

	shadow := ControllerSet{
		Controllers: make([]*Controller, len(cc.Controllers)),
		LAN:         &LAN{},
	}

	lan := cc.LAN.clone()
	shadow.LAN = &lan

	for k, v := range cc.Controllers {
		shadow.Controllers[k] = v.clone()
	}

	return shadow
}

func Export(file string, controllers []*Controller, doors map[catalog.OID]doors.Door) error {
	guard.RLock()

	defer guard.RUnlock()

	conf := config.NewConfig()
	if err := conf.Load(file); err != nil {
		return err
	}

	devices := config.DeviceMap{}
	for _, c := range controllers {
		if c.deviceID != nil && *c.deviceID != 0 && c.deleted == nil {
			device := config.Device{
				Name:     c.name,
				Address:  nil,
				Doors:    []string{"", "", "", ""},
				TimeZone: "",
				Rollover: 100000,
			}

			if c.IP != nil {
				device.Address = (*net.UDPAddr)(c.IP)
			}

			if c.TimeZone != nil {
				device.TimeZone = *c.TimeZone
			}

			if d, ok := doors[c.Doors[1]]; ok {
				device.Doors[0] = d.Name
			}

			if d, ok := doors[c.Doors[2]]; ok {
				device.Doors[1] = d.Name
			}

			if d, ok := doors[c.Doors[3]]; ok {
				device.Doors[2] = d.Name
			}

			if d, ok := doors[c.Doors[4]]; ok {
				device.Doors[3] = d.Name
			}

			devices[*c.deviceID] = &device
		}
	}

	conf.Devices = devices

	var b bytes.Buffer
	conf.Write(&b)

	tmp, err := ioutil.TempFile(os.TempDir(), "uhppoted.conf_")
	if err != nil {
		return err
	}

	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(b.Bytes()); err != nil {
		return err
	}

	if err := tmp.Close(); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(file), 0770); err != nil {
		return err
	}

	return os.Rename(tmp.Name(), file)
}

func (cc *ControllerSet) Sync() []catalog.Object {
	objects := []catalog.Object{}

	if list := cc.LAN.synchTime(cc.Controllers); list != nil {
		objects = append(objects, list...)
	}

	if list := cc.LAN.synchDoors(cc.Controllers); list != nil {
		objects = append(objects, list...)
	}

	return objects
}

func (cc *ControllerSet) Compare(permissions acl.ACL) error {
	return cc.LAN.compareACL(cc.Controllers, permissions)
}

func (cc *ControllerSet) UpdateACL(acl acl.ACL) {
	cc.LAN.updateACL(cc.Controllers, acl)
}

func (cc *ControllerSet) Validate() error {
	if cc != nil {
		return validate(*cc)
	}

	return nil
}

func (cc *ControllerSet) add(auth auth.OpAuth, c Controller) (*Controller, error) {
	id := uint32(0)
	if c.deviceID != nil {
		id = *c.deviceID
	}

	record := c.clone()
	record.oid = catalog.OID(catalog.NewController(id))
	record.created = types.DateTime(time.Now())

	if auth != nil {
		if err := auth.CanAddController(record); err != nil {
			return nil, err
		}
	}

	cc.Controllers = append(cc.Controllers, record)

	return record, nil
}

func validate(cc ControllerSet) error {
	devices := map[uint32]string{}

	for _, c := range cc.Controllers {
		OID := c.OID()
		if OID == "" {
			return fmt.Errorf("Invalid controller OID (%v)", OID)
		}

		if c.deleted != nil {
			continue
		}

		if c.deviceID != nil && *c.deviceID != 0 {
			id := *c.deviceID

			if _, ok := devices[id]; ok {
				return fmt.Errorf("Duplicate controller ID (%v)", id)
			}

			devices[id] = string(OID)
		}
	}

	return nil
}

func scrub(cc *ControllerSet) error {
	return nil
}

func warn(err error) {
	log.Printf("ERROR %v", err)
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
