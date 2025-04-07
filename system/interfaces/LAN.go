package interfaces

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	lib "github.com/uhppoted/uhppote-core/types"

	"github.com/uhppoted/uhppote-core/uhppote"
	"github.com/uhppoted/uhppoted-lib/acl"
	"github.com/uhppoted/uhppoted-lib/uhppoted"

	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/log"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	"github.com/uhppoted/uhppoted-httpd/system/db"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type LAN struct {
	catalog.CatalogInterface
	Name             string
	BindAddress      lib.BindAddr
	BroadcastAddress lib.BroadcastAddr
	ListenAddress    lib.ListenAddr
	Debug            bool

	ch       chan types.EventsList
	created  types.Timestamp
	modified types.Timestamp
	deleted  types.Timestamp

	unconfigured bool
}

type kv = struct {
	field schema.Suffix
	value interface{}
}

var created = types.TimestampNow()

const MAX = 5

func (l LAN) String() string {
	return fmt.Sprintf("%v", l.Name)
}

func (l LAN) IsValid() bool {
	return l.validate() == nil
}

func (l LAN) validate() error {
	if strings.TrimSpace(l.Name) == "" {
		return fmt.Errorf("LAN name is blank")
	}

	return nil
}

func (l LAN) IsDeleted() bool {
	return !l.deleted.IsZero()
}

func (l *LAN) AsObjects(a *auth.Authorizator) []schema.Object {
	list := []kv{}

	if l.IsDeleted() {
		list = append(list, kv{LANDeleted, l.deleted})
	} else {
		list = append(list, kv{LANStatus, l.status()})
		list = append(list, kv{LANCreated, l.created})
		list = append(list, kv{LANDeleted, l.deleted})
		list = append(list, kv{LANType, "LAN"})
		list = append(list, kv{LANName, l.Name})
		list = append(list, kv{LANBindAddress, l.BindAddress})
		list = append(list, kv{LANBroadcastAddress, l.BroadcastAddress})
		list = append(list, kv{LANListenAddress, l.ListenAddress})
	}

	return l.toObjects(list, a)
}

func (l LAN) AsRuleEntity() (string, interface{}) {
	entity := struct {
		Type string
		Name string
	}{
		Type: "LAN",
		Name: fmt.Sprintf("%v", l.Name),
	}

	return "lan", &entity
}

func (l LAN) CacheKey() string {
	return ""
}

func (l *LAN) set(a *auth.Authorizator, oid schema.OID, value string, dbc db.DBC) ([]schema.Object, error) {
	if l == nil {
		return []schema.Object{}, nil
	}

	if l.IsDeleted() {
		return l.toObjects([]kv{{LANDeleted, l.deleted}}, a), fmt.Errorf("LAN has been deleted")
	}

	uid := auth.UID(a)
	list := []kv{}

	switch oid {
	case l.OID.Append(LANName):
		if err := CanUpdate(a, l, "name", value); err != nil {
			return nil, err
		} else {
			l.log(dbc, uid, "update", l.OID, "name", l.Name, value, "Updated name from %v to %v", l.Name, value)

			l.Name = value
			l.modified = types.TimestampNow()

			list = append(list, kv{LANName, l.Name})
		}

	case l.OID.Append(LANBindAddress):
		if addr, err := lib.ParseBindAddr(value); err != nil {
			return nil, err
		} else if !addr.IsValid() {
			return nil, fmt.Errorf("invalid bind address (%v)", value)
		} else if err := CanUpdate(a, l, "bind", addr); err != nil {
			return nil, err
		} else {
			l.log(dbc, uid, "update", l.OID, "bind", l.BindAddress, value, "Updated bind address from %v to %v", l.BindAddress, value)

			l.BindAddress = addr
			l.modified = types.TimestampNow()

			list = append(list, kv{LANBindAddress, l.BindAddress})
		}

	case l.OID.Append(LANBroadcastAddress):
		if addr, err := lib.ParseBroadcastAddr(value); err != nil {
			return nil, err
		} else if !addr.IsValid() {
			return nil, fmt.Errorf("invalid broadcast address (%v)", value)
		} else if err := CanUpdate(a, l, "broadcast", addr); err != nil {
			return nil, err
		} else {
			l.log(dbc, uid, "update", l.OID, "broadcast", l.BroadcastAddress, value, "Updated broadcast address from %v to %v", l.BroadcastAddress, value)

			l.BroadcastAddress = addr
			l.modified = types.TimestampNow()

			list = append(list, kv{LANBroadcastAddress, l.BroadcastAddress})
		}

	case l.OID.Append(LANListenAddress):
		if addr, err := lib.ParseListenAddr(value); err != nil {
			return nil, err
		} else if err = CanUpdate(a, l, "listen", addr); err != nil {
			return nil, err
		} else {
			l.log(dbc, uid, "update", l.OID, "listen", l.ListenAddress, value, "Updated listen address from %v to %v", l.ListenAddress, value)

			l.ListenAddress = addr
			l.modified = types.TimestampNow()

			list = append(list, kv{LANListenAddress, l.ListenAddress})
		}
	}

	list = append(list, kv{LANStatus, l.status()})

	return l.toObjects(list, a), nil
}

func (l LAN) toObjects(list []kv, a *auth.Authorizator) []schema.Object {
	objects := []schema.Object{}

	if err := CanView(a, l, "OID", l.OID); err == nil && !l.IsDeleted() {
		objects = append(objects, catalog.NewObject(l.OID, ""))
	}

	for _, v := range list {
		field := lookup[v.field]
		if err := CanView(a, l, field, v.value); err == nil {
			objects = append(objects, catalog.NewObject2(l.OID, v.field, v.value))
		}
	}

	return objects
}

func (l LAN) Clone() LAN {
	return LAN{
		CatalogInterface: catalog.CatalogInterface{
			OID: l.OID,
		},
		Name:             l.Name,
		BindAddress:      l.BindAddress,
		BroadcastAddress: l.BroadcastAddress,
		ListenAddress:    l.ListenAddress,
		Debug:            l.Debug,

		created:      l.created,
		modified:     l.modified,
		deleted:      l.deleted,
		unconfigured: l.unconfigured,
	}
}

func (l *LAN) search(controllers []types.IController) ([]uint32, error) {
	list := []uint32{}

	api := l.api(controllers)
	if devices, err := api.GetDevices(uhppoted.GetDevicesRequest{}); err != nil {
		return list, err
	} else if devices == nil {
		return list, fmt.Errorf("got %v response to get-devices request", devices)
	} else {
		for k := range devices.Devices {
			list = append(list, k)
		}
	}

	return list, nil
}

// A long-running function i.e. expects to be invoked from an external goroutine
func (l *LAN) refresh(c types.IController) {
	log.Infof("%v: refreshing LAN controller status", c.ID())

	api := l.api([]types.IController{c})
	deviceIDu := c.ID()
	deviceID := uhppoted.DeviceID(c.ID())

	if info, err := api.GetDevice(uhppoted.GetDeviceRequest{DeviceID: deviceID}); err != nil {
		log.Warnf("%v", err)
	} else if info == nil {
		log.Warnf("<%v> response to get-device request for %v", info, deviceID)
	} else {
		catalog.PutV(c.OID(), ControllerTouched, time.Now())
		catalog.PutV(c.OID(), ControllerEndpointAddress, lib.ControllerAddrFrom(info.Address, 60000))
	}

	if status, err := api.UHPPOTE.GetStatus(deviceIDu); err != nil {
		log.Warnf("%v", err)
	} else if status == nil {
		log.Warnf("<%v> response to get-status request for %v", status, deviceID)
	} else {
		catalog.PutV(c.OID(), ControllerTouched, time.Now())
		catalog.PutV(c.OID(), ControllerDateTimeCurrent, status.SystemDateTime)
	}

	if cards, err := api.GetCardRecords(uhppoted.GetCardRecordsRequest{DeviceID: deviceID}); err != nil {
		log.Warnf("%v", err)
	} else if cards == nil {
		log.Warnf("<%v> response to get-card-records request for %v", cards, deviceID)
	} else {
		catalog.PutV(c.OID(), ControllerTouched, time.Now())
		catalog.PutV(c.OID(), ControllerCardsCount, cards.Cards)
	}

	for _, d := range []uint8{1, 2, 3, 4} {
		if delay, err := api.GetDoorDelay(uhppoted.GetDoorDelayRequest{DeviceID: deviceID, Door: d}); err != nil {
			log.Warnf("%v", err)
		} else if delay == nil {
			log.Warnf("<%v> response to get-door-delay request for %v", delay, deviceID)
		} else if door, ok := c.Door(delay.Door); ok {
			catalog.PutV(door, DoorDelay, delay.Delay)
		}
	}

	for _, d := range []uint8{1, 2, 3, 4} {
		if control, err := api.GetDoorControl(uhppoted.GetDoorControlRequest{DeviceID: deviceID, Door: d}); err != nil {
			log.Warnf("%v", err)
		} else if control == nil {
			log.Warnf("<%v> response to get-door-control request for %v", control, deviceID)
		} else if door, ok := c.Door(control.Door); ok {
			catalog.PutV(door, DoorControl, control.Control)
		}
	}

	if antipassback, err := api.UHPPOTE.GetAntiPassback(deviceIDu); err != nil {
		log.Warnf("%v", err)
	} else {
		catalog.PutV(c.OID(), ControllerTouched, time.Now())
		catalog.PutV(c.OID(), ControllerAntiPassback, antipassback)
	}
}

func (l *LAN) getEvents(c types.IController, intervals []types.Interval) {
	api := l.api([]types.IController{c})
	deviceID := c.ID()
	oid := c.OID()

	log.Infof("%v: retrieving LAN controller events (%v)", deviceID, intervals)

	if first, last, current, err := api.GetEventIndices(deviceID); err != nil {
		log.Warnf("%v", err)
	} else {
		catalog.PutV(oid, ControllerTouched, time.Now())
		catalog.PutV(oid, ControllerEventsFirst, first)
		catalog.PutV(oid, ControllerEventsLast, last)
		catalog.PutV(oid, ControllerEventsCurrent, current)

		status := types.StatusOk
		for _, interval := range intervals {
			if interval.Contains(last) || interval.Contains(first) || (interval.From >= first && interval.To <= last) {
				status = types.StatusIncomplete
				break
			}
		}

		catalog.PutV(oid, ControllerEventsStatus, status)

		count := 0
		events := []uhppoted.Event{}

		f := func(index uint32) {
			if e, err := api.UHPPOTE.GetEvent(deviceID, index); err != nil {
				log.Warnf("%v", err)
			} else if e == nil {
				log.Warnf("%v: missing event %v", deviceID, index)
				events = append(events, uhppoted.Event{
					DeviceID: deviceID,
					Index:    index,
				})
			} else {
				events = append(events, uhppoted.Event{
					DeviceID:   deviceID,
					Index:      e.Index,
					Type:       e.Type,
					Granted:    e.Granted,
					Door:       e.Door,
					Direction:  e.Direction,
					CardNumber: e.CardNumber,
					Timestamp:  e.Timestamp,
					Reason:     e.Reason,
				})
			}

			count++
		}

		for _, interval := range intervals {
			if interval.Contains(last) {
				index := interval.From
				if index < first {
					index = first
				}

				for index <= last && count < MAX {
					f(index)
					index++
				}
			}

			if interval.Contains(first) {
				index := interval.To
				if index > last {
					index = last
				}

				for index >= first && count < MAX {
					f(index)
					index--
				}
			}

			if interval.From >= first && interval.To <= last {
				for index := interval.From; index <= interval.To; index++ {
					f(index)
				}
			}
		}

		if l.ch != nil {
			l.ch <- types.EventsList{
				DeviceID: deviceID,
				Events:   events,
			}
		}
	}
}

func (l *LAN) setTime(c types.IController, t time.Time) {
	lock(c.ID())
	defer unlock(c.ID())

	api := l.api([]types.IController{c})
	deviceID := c.ID()
	location := c.TimeZone()
	datetime := time.Time(t.In(location))

	if response, err := api.UHPPOTE.SetTime(deviceID, datetime); err != nil {
		log.Warnf("%v", err)
	} else if response != nil {
		catalog.PutV(c.OID(), ControllerDateTimeModified, false)

		if status, err := api.UHPPOTE.GetStatus(deviceID); err != nil {
			log.Warnf("%v", err)
		} else if status == nil {
			log.Warnf("Got %v response to get-status request for %v", status, deviceID)
		} else {
			catalog.PutV(c.OID(), ControllerDateTimeCurrent, status.SystemDateTime)
		}

		log.Infof("%v  set date/time: %v", deviceID, response.DateTime)
	}
}

func (l *LAN) setDoor(c types.IController, door uint8, mode lib.ControlState, delay uint8) error {
	lock(c.ID())
	defer unlock(c.ID())

	api := l.api([]types.IController{c})
	deviceID := c.ID()

	if state, err := api.UHPPOTE.GetDoorControlState(deviceID, door); err != nil {
		return err
	} else if state == nil {
		return fmt.Errorf("got %v response to get-door request for %v", state, deviceID)
	} else {
		m := mode
		d := delay

		if m == lib.ModeUnknown {
			m = state.ControlState
		}

		if d == 0 {
			d = state.Delay
		}

		if _, err := api.UHPPOTE.SetDoorControlState(deviceID, door, m, d); err != nil {
			return err
		}

		log.Infof("%v  set door %v mode:%-15v delay:%vs", deviceID, door, m, d)

		return nil
	}
}

func (l *LAN) setInterlock(c types.IController, interlock lib.Interlock) error {
	lock(c.ID())
	defer unlock(c.ID())

	api := l.api([]types.IController{c})
	deviceID := c.ID()

	if ok, err := api.UHPPOTE.SetInterlock(deviceID, interlock); err != nil {
		return err
	} else if !ok {
		return fmt.Errorf("%v  failed to set door interlock mode (%v)", deviceID, interlock)
	} else {
		return nil
	}
}

func (l *LAN) setAntiPassback(c types.IController, antipassback lib.AntiPassback) error {
	lock(c.ID())
	defer unlock(c.ID())

	api := l.api([]types.IController{c})
	deviceID := c.ID()

	if ok, err := api.UHPPOTE.SetAntiPassback(deviceID, antipassback); err != nil {
		return err
	} else if !ok {
		return fmt.Errorf("%v  failed to set anti-passback (%v)", deviceID, antipassback)
	} else {
		return nil
	}
}

func (l *LAN) activateKeypads(c types.IController, keypads map[uint8]bool) error {
	lock(c.ID())
	defer unlock(c.ID())

	api := l.api([]types.IController{c})
	deviceID := c.ID()

	if ok, err := api.UHPPOTE.ActivateKeypads(deviceID, keypads); err != nil {
		return err
	} else if !ok {
		return fmt.Errorf("%v  failed to activate/deactivate keypads (%v)", deviceID, keypads)
	} else {
		return nil
	}
}

func (l *LAN) setDoorPasscodes(c types.IController, door uint8, passcodes ...uint32) error {
	lock(c.ID())
	defer unlock(c.ID())

	api := l.api([]types.IController{c})
	deviceID := c.ID()

	if ok, err := api.UHPPOTE.SetDoorPasscodes(deviceID, door, passcodes...); err != nil {
		return err
	} else if !ok {
		return fmt.Errorf("%v  failed to set door %v passcodes", deviceID, door)
	} else {
		return nil
	}
}

func (l *LAN) putCard(c types.IController, cardID uint32, PIN uint32, from, to lib.Date, permissions map[uint8]uint8) {
	lock(c.ID())
	defer unlock(c.ID())

	api := l.api([]types.IController{c})
	deviceID := c.ID()

	card := lib.Card{
		CardNumber: cardID,
		PIN:        lib.PIN(PIN),
		From:       from,
		To:         to,
		Doors: map[uint8]uint8{
			1: permissions[1],
			2: permissions[2],
			3: permissions[3],
			4: permissions[4],
		},
	}

	if ok, err := api.UHPPOTE.PutCard(deviceID, card); err != nil {
		log.Warnf("%v", err)
	} else if !ok {
		log.Warnf("%v", fmt.Errorf("%v  failed to update card %v", deviceID, cardID))
	} else {
		log.Infof("%v  put card %v", deviceID, card)
	}
}

func (l *LAN) deleteCard(c types.IController, card uint32) {
	lock(c.ID())
	defer unlock(c.ID())

	api := l.api([]types.IController{c})
	deviceID := c.ID()

	if ok, err := api.UHPPOTE.DeleteCard(deviceID, card); err != nil {
		log.Warnf("%v", err)
	} else if !ok {
		log.Warnf("%v", fmt.Errorf("%v  failed to delete card %v", deviceID, card))
	} else {
		log.Infof("%v  deleted card %v", deviceID, card)
	}
}

func (l *LAN) compareACL(controllers []types.IController, permissions acl.ACL, withPIN bool) (map[uint32]acl.Diff, error) {
	log.Debugf("Comparing ACL (with-pin:%v)", withPIN)

	devices := []uhppote.Device{}
	api := l.api(controllers)
	for _, v := range api.UHPPOTE.DeviceList() {
		device := v
		devices = append(devices, device)
	}

	current, errors := acl.GetACL(api.UHPPOTE, devices)
	for _, err := range errors {
		log.Warnf("%v", err)
	}

	f := func(permissions, current acl.ACL) (map[uint32]acl.Diff, error) {
		if withPIN {
			return acl.CompareWithPIN(permissions, current)
		} else {
			return acl.Compare(permissions, current)
		}
	}

	compare, err := f(permissions, current)
	if err != nil {
		return nil, err
	} else if compare == nil {
		return nil, fmt.Errorf("invalid ACL compare report: %v", compare)
	}

	for k, v := range compare {
		log.Infof("ACL %v  unchanged:%-3v updated:%-3v added:%-3v deleted:%-3v", k, len(v.Unchanged), len(v.Updated), len(v.Added), len(v.Deleted))
	}

	diff := acl.SystemDiff(compare)
	report := diff.Consolidate()
	if report == nil {
		return nil, fmt.Errorf("invalid consolidated ACL compare report: %v", report)
	}

	unchanged := len(report.Unchanged)
	updated := len(report.Updated)
	added := len(report.Added)
	deleted := len(report.Deleted)

	log.Infof("ACL compare    unchanged:%-3v updated:%-3v added:%-3v deleted:%-3v", unchanged, updated, added, deleted)

	for _, c := range controllers {
		for _, d := range devices {
			if c.ID() == d.DeviceID {
				rs := compare[c.ID()]
				if len(rs.Updated)+len(rs.Added)+len(rs.Deleted) > 0 {
					catalog.PutV(c.OID(), ControllerCardsStatus, types.StatusError)
				} else {
					catalog.PutV(c.OID(), ControllerCardsStatus, types.StatusOk)
				}
				break
			}
		}
	}

	return compare, nil
}

func (l *LAN) status() types.Status {
	return types.StatusOk
}

func (l *LAN) api(controllers []types.IController) *uhppoted.UHPPOTED {
	devices := []uhppote.Device{}

	for _, v := range controllers {
		name := v.Name()
		id := v.ID()
		addr := lib.ControllerAddrFrom(v.EndPoint().Addr(), v.EndPoint().Port())
		doors := []string{}
		tz := v.TimeZone()
		protocol := v.Protocol()

		// NTS: 'found' controllers will have a zero value address (only configured controllers have an address)
		if device := uhppote.NewDevice(name, id, addr, protocol, doors, tz); device.IsValid() {
			devices = append(devices, device)
		}
	}

	u := uhppote.NewUHPPOTE(l.BindAddress, l.BroadcastAddress, l.ListenAddress, 1*time.Second, devices, l.Debug)
	api := uhppoted.UHPPOTED{
		UHPPOTE: u,
	}

	return &api
}

func (l LAN) serialize() ([]byte, error) {
	record := struct {
		OID              schema.OID        `json:"OID"`
		Name             string            `json:"name,omitempty"`
		BindAddress      lib.BindAddr      `json:"bind-address,omitempty"`
		BroadcastAddress lib.BroadcastAddr `json:"broadcast-address,omitempty"`
		ListenAddress    lib.ListenAddr    `json:"listen-address,omitempty"`
		Created          types.Timestamp   `json:"created,omitempty"`
		Modified         types.Timestamp   `json:"modified,omitempty"`
	}{
		OID:              l.OID,
		Name:             l.Name,
		BindAddress:      l.BindAddress,
		BroadcastAddress: l.BroadcastAddress,
		ListenAddress:    l.ListenAddress,
		Created:          l.created.UTC(),
		Modified:         l.modified.UTC(),
	}

	return json.MarshalIndent(record, "", "  ")
}

func (l *LAN) deserialize(bytes []byte) error {
	created = created.Add(1 * time.Minute)

	record := struct {
		OID              schema.OID        `json:"OID"`
		Name             string            `json:"name,omitempty"`
		BindAddress      lib.BindAddr      `json:"bind-address,omitempty"`
		BroadcastAddress lib.BroadcastAddr `json:"broadcast-address,omitempty"`
		ListenAddress    lib.ListenAddr    `json:"listen-address,omitempty"`
		Created          types.Timestamp   `json:"created,omitempty"`
		Modified         types.Timestamp   `json:"modified,omitempty"`
	}{
		Created: created,
	}

	if err := json.Unmarshal(bytes, &record); err != nil {
		return err
	}

	l.OID = record.OID
	l.Name = record.Name
	l.BindAddress = record.BindAddress
	l.BroadcastAddress = record.BroadcastAddress
	l.ListenAddress = record.ListenAddress
	l.created = record.Created
	l.modified = record.Modified
	l.unconfigured = false

	return nil
}

func (l *LAN) log(dbc db.DBC, uid string, op string, OID schema.OID, field string, before, after any, format string, fields ...any) {
	dbc.Log(uid, op, OID, "interface", "LAN", l.Name, field, before, after, format, fields...)
}
