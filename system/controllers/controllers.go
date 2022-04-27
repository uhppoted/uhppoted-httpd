package controllers

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/uhppoted/uhppoted-lib/acl"

	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/log"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	"github.com/uhppoted/uhppoted-httpd/system/db"
	"github.com/uhppoted/uhppoted-httpd/system/interfaces"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type Controllers struct {
	controllers []*Controller
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

func NewControllers() Controllers {
	return Controllers{
		controllers: []*Controller{},
	}
}

func (cc *Controllers) AsObjects(a *auth.Authorizator) []schema.Object {
	objects := []schema.Object{}

	for _, c := range cc.controllers {
		if c.IsValid() {
			catalog.Join(&objects, c.AsObjects(a)...)
		}
	}

	return objects
}

func (cc *Controllers) AsIControllers() []types.IController {
	list := []types.IController{}

	for _, c := range cc.controllers {
		if c.DeviceID != 0 && !c.IsDeleted() {
			list = append(list, c.AsIController())
		}
	}

	return list
}

func (cc *Controllers) UpdateByOID(a *auth.Authorizator, oid schema.OID, value string, dbc db.DBC) ([]schema.Object, error) {
	if cc == nil {
		return nil, nil
	}

	uid := auth.UID(a)

	for _, c := range cc.controllers {
		if c != nil && c.OID.Contains(oid) {
			return c.set(a, oid, value, dbc)
		}
	}

	objects := []schema.Object{}

	if oid == "<new>" {
		if c, err := cc.add(a, Controller{}); err != nil {
			return nil, err
		} else if c == nil {
			return nil, fmt.Errorf("Failed to add 'new' controller")
		} else {
			OID := c.OID

			catalog.Join(&objects, catalog.NewObject(OID, "new"))
			catalog.Join(&objects, catalog.NewObject2(OID, ControllerStatus, "new"))
			catalog.Join(&objects, catalog.NewObject2(OID, ControllerCreated, c.created))

			c.log(uid, "add", OID, "controller", fmt.Sprintf("Added 'new' controller"), "", "", dbc)
		}
	}

	return objects, nil
}

func (cc *Controllers) DeleteByOID(a *auth.Authorizator, oid schema.OID, dbc db.DBC) ([]schema.Object, error) {
	objects := []schema.Object{}

	if cc != nil {
		for _, c := range cc.controllers {
			if c != nil && c.OID == oid {
				return c.delete(a, dbc)
			}
		}
	}

	return objects, nil
}

func (cc *Controllers) List() []Controller {
	list := []Controller{}

	for _, c := range cc.controllers {
		if c != nil {
			list = append(list, *c)
		}
	}

	return list
}

func (cc *Controllers) Load(blob json.RawMessage) error {
	rs := []json.RawMessage{}
	if err := json.Unmarshal(blob, &rs); err != nil {
		return err
	}

	cc.controllers = []*Controller{}
	for _, v := range rs {
		var c Controller
		if err := c.deserialize(v); err != nil {
			log.Warnf("%v", err)
		} else {
			cc.controllers = append(cc.controllers, &c)
		}
	}

	for _, c := range cc.controllers {
		oid := c.OID
		catalog.PutT(c.CatalogController)
		catalog.PutV(oid, ControllerName, c.name)
		catalog.PutV(oid, ControllerDeviceID, c.DeviceID)
		catalog.PutV(oid, ControllerDateTimeModified, false)
		catalog.PutV(oid, ControllerDoor1, c.doors[1])
		catalog.PutV(oid, ControllerDoor2, c.doors[2])
		catalog.PutV(oid, ControllerDoor3, c.doors[3])
		catalog.PutV(oid, ControllerDoor4, c.doors[4])
	}

	return nil
}

func (cc Controllers) Save() (json.RawMessage, error) {
	if err := cc.Validate(); err != nil {
		return nil, err
	}

	serializable := []json.RawMessage{}
	for _, c := range cc.controllers {
		if bytes, err := c.serialize(); err == nil && bytes != nil {
			serializable = append(serializable, bytes)
		}
	}

	return json.MarshalIndent(serializable, "", "  ")
}

func (cc *Controllers) Sweep(retention time.Duration) {
	if cc == nil {
		return
	}

	cutoff := time.Now().Add(-retention)
	for i, v := range cc.controllers {
		if v.IsDeleted() && v.deleted.Before(cutoff) {
			cc.controllers = append(cc.controllers[:i], cc.controllers[i+1:]...)
		}
	}
}

func (cc Controllers) Print() {
	serializable := []json.RawMessage{}
	for _, c := range cc.controllers {
		if bytes, err := c.serialize(); err == nil && bytes != nil {
			serializable = append(serializable, bytes)
		}
	}

	if b, err := json.MarshalIndent(serializable, "", "  "); err == nil {
		fmt.Printf("----------------- CONTROLLERS\n%s\n", string(b))
	}
}

func (cc *Controllers) Found(found []uint32) {
loop:
	for _, v := range found {
		for _, c := range cc.controllers {
			if c.DeviceID == v && !c.IsDeleted() {
				continue loop
			}
		}

		log.Infof("Adding unconfigured controller %v", v)

		id := v // because .. Go loop variable gotcha (the loop variable is mutable)
		c := Controller{
			CatalogController: catalog.CatalogController{
				DeviceID: id,
			},
			created: types.TimestampNow(),
		}

		c.OID = catalog.NewT(c.CatalogController)

		cc.controllers = append(cc.controllers, &c)
	}
}

// NTS: 'added' is specifically not cloned - it has a lifetime for the duration of
//      the 'shadow' copy only
func (cc *Controllers) Clone() Controllers {
	guard.RLock()
	defer guard.RUnlock()

	shadow := Controllers{
		controllers: make([]*Controller, len(cc.controllers)),
	}

	for k, v := range cc.controllers {
		shadow.controllers[k] = v.clone()
	}

	return shadow
}

func (cc *Controllers) CompareACL(i interfaces.Interfaces, permissions acl.ACL) error {
	var lan LAN

	if v, ok := i.LAN(); !ok {
		return fmt.Errorf("No active LAN subsystem")
	} else {
		lan = LAN{
			interfaces: i,
			lan:        v,
		}
	}

	return lan.compare(cc.controllers, permissions)
}

func (cc *Controllers) UpdateACL(i interfaces.Interfaces, permissions acl.ACL) error {
	var lan LAN

	if v, ok := i.LAN(); !ok {
		return fmt.Errorf("No active LAN subsystem")
	} else {
		lan = LAN{
			interfaces: i,
			lan:        v,
		}
	}

	return lan.update(cc.controllers, permissions)
}

func (cc Controllers) Validate() error {
	devices := map[uint32]string{}

	for _, c := range cc.controllers {
		OID := c.OID
		if OID == "" {
			return fmt.Errorf("Invalid controller OID (%v)", OID)
		}

		if c.IsDeleted() {
			continue
		}

		if err := c.validate(); err != nil {
			if !c.modified.IsZero() {
				return err
			}
		}

		if c.DeviceID != 0 {
			if _, ok := devices[c.DeviceID]; ok {
				return fmt.Errorf("Duplicate controller ID (%v)", c.DeviceID)
			}

			devices[c.DeviceID] = string(OID)
		}
	}

	return nil
}

func (cc *Controllers) add(a auth.OpAuth, c Controller) (*Controller, error) {
	controller := c.clone()
	controller.OID = schema.OID(catalog.NewT(c.CatalogController))
	controller.created = types.TimestampNow()

	if a != nil {
		if err := a.CanAdd(controller, auth.Controllers); err != nil {
			return nil, err
		}
	}

	cc.controllers = append(cc.controllers, controller)

	return controller, nil
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
