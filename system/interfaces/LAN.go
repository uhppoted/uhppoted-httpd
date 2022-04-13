package interfaces

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	core "github.com/uhppoted/uhppote-core/types"

	"github.com/uhppoted/uhppote-core/uhppote"
	"github.com/uhppoted/uhppoted-lib/acl"
	"github.com/uhppoted/uhppoted-lib/uhppoted"

	"github.com/uhppoted/uhppoted-httpd/audit"
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
	BindAddress      core.BindAddr
	BroadcastAddress core.BroadcastAddr
	ListenAddress    core.ListenAddr
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

func (l *LAN) AsRuleEntity() (string, interface{}) {
	entity := struct {
		Type string
		Name string
	}{}

	if l != nil {
		entity.Type = "LAN"
		entity.Name = fmt.Sprintf("%v", l.Name)
	}

	return "lan", &entity
}

func (l *LAN) set(a *auth.Authorizator, oid schema.OID, value string, dbc db.DBC) ([]schema.Object, error) {
	if l == nil {
		return []schema.Object{}, nil
	}

	if l.IsDeleted() {
		return l.toObjects([]kv{{LANDeleted, l.deleted}}, a), fmt.Errorf("LAN has been deleted")
	}

	f := func(field string, value interface{}) error {
		if a != nil {
			return a.CanUpdate(l, field, value, auth.Interfaces)
		}

		return nil
	}

	uid := auth.UID(a)
	list := []kv{}

	switch oid {
	case l.OID.Append(LANName):
		if err := f("name", value); err != nil {
			return nil, err
		} else {
			l.log(uid,
				"update",
				l.OID,
				"name",
				fmt.Sprintf("Updated name from %v to %v", stringify(l.Name, BLANK), stringify(value, BLANK)),
				stringify(l.Name, BLANK),
				stringify(value, BLANK),
				dbc)

			l.Name = value
			l.modified = types.TimestampNow()

			list = append(list, kv{LANName, l.Name})
		}

	case l.OID.Append(LANBindAddress):
		if addr, err := core.ResolveBindAddr(value); err != nil {
			return nil, err
		} else if err := f("bind", addr); err != nil {
			return nil, err
		} else {
			l.log(uid,
				"update",
				l.OID,
				"bind",
				fmt.Sprintf("Updated bind address from %v to %v", stringify(l.BindAddress, BLANK), stringify(value, BLANK)),
				stringify(l.BindAddress, BLANK),
				stringify(value, BLANK),
				dbc)

			l.BindAddress = *addr
			l.modified = types.TimestampNow()

			list = append(list, kv{LANBindAddress, l.BindAddress})
		}

	case l.OID.Append(LANBroadcastAddress):
		if addr, err := core.ResolveBroadcastAddr(value); err != nil {
			return nil, err
		} else if err := f("broadcast", addr); err != nil {
			return nil, err
		} else {
			l.log(uid,
				"update",
				l.OID,
				"broadcast",
				fmt.Sprintf("Updated broadcast address from %v to %v", stringify(l.BroadcastAddress, BLANK), stringify(value, BLANK)),
				stringify(l.BroadcastAddress, BLANK),
				stringify(value, BLANK),
				dbc)

			l.BroadcastAddress = *addr
			l.modified = types.TimestampNow()

			list = append(list, kv{LANBroadcastAddress, l.BroadcastAddress})
		}

	case l.OID.Append(LANListenAddress):
		if addr, err := core.ResolveListenAddr(value); err != nil {
			return nil, err
		} else if err = f("listen", addr); err != nil {
			return nil, err
		} else {
			l.log(uid,
				"update",
				l.OID,
				"listen",
				fmt.Sprintf("Updated listen address from %v to %v", stringify(l.ListenAddress, BLANK), stringify(value, BLANK)),
				stringify(l.ListenAddress, BLANK),
				stringify(value, BLANK),
				dbc)

			l.ListenAddress = *addr
			l.modified = types.TimestampNow()

			list = append(list, kv{LANListenAddress, l.ListenAddress})
		}
	}

	list = append(list, kv{LANStatus, l.status()})

	return l.toObjects(list, a), nil
}

func (l *LAN) toObjects(list []kv, a *auth.Authorizator) []schema.Object {
	f := func(l *LAN, field string, value interface{}) bool {
		if a != nil {
			if err := a.CanView(l, field, value, auth.Interfaces); err != nil {
				return false
			}
		}

		return true
	}

	objects := []schema.Object{}

	if !l.IsDeleted() && f(l, "OID", l.OID) {
		objects = append(objects, catalog.NewObject(l.OID, ""))
	}

	for _, v := range list {
		field, _ := lookup[v.field]
		if f(l, field, v.value) {
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

func (l *LAN) Search(controllers []Controller) ([]uint32, error) {
	list := []uint32{}

	api := l.api(controllers)
	if devices, err := api.GetDevices(uhppoted.GetDevicesRequest{}); err != nil {
		return list, err
	} else if devices == nil {
		return list, fmt.Errorf("Got %v response to get-devices request", devices)
	} else {
		for k, _ := range devices.Devices {
			list = append(list, k)
		}
	}

	return list, nil
}

// A long-running function i.e. expects to be invoked from an external goroutine
func (l *LAN) Refresh(c Controller) {
	log.Infof("%v: refreshing LAN controller status", c.ID())

	api := l.api([]Controller{c})
	deviceID := uhppoted.DeviceID(c.ID())

	if info, err := api.GetDevice(uhppoted.GetDeviceRequest{DeviceID: deviceID}); err != nil {
		log.Warnf("%v", err)
	} else if info == nil {
		log.Warnf("Got %v response to get-device request for %v", info, deviceID)
	} else {
		catalog.PutV(c.OIDx(), ControllerTouched, time.Now())
		catalog.PutV(c.OIDx(), ControllerEndpointAddress, core.Address(info.Address))
	}

	if status, err := api.GetStatus(uhppoted.GetStatusRequest{DeviceID: deviceID}); err != nil {
		log.Warnf("%v", err)
	} else if status == nil {
		log.Warnf("Got %v response to get-status request for %v", status, deviceID)
	} else {
		catalog.PutV(c.OIDx(), ControllerTouched, time.Now())
		catalog.PutV(c.OIDx(), ControllerDateTimeCurrent, status.Status.SystemDateTime)
	}

	if cards, err := api.GetCardRecords(uhppoted.GetCardRecordsRequest{DeviceID: deviceID}); err != nil {
		log.Warnf("%v", err)
	} else if cards == nil {
		log.Warnf("Got %v response to get-card-records request for %v", cards, deviceID)
	} else {
		catalog.PutV(c.OIDx(), ControllerTouched, time.Now())
		catalog.PutV(c.OIDx(), ControllerCardsCount, cards.Cards)
	}

	for _, d := range []uint8{1, 2, 3, 4} {
		if delay, err := api.GetDoorDelay(uhppoted.GetDoorDelayRequest{DeviceID: deviceID, Door: d}); err != nil {
			log.Warnf("%v", err)
		} else if delay == nil {
			log.Warnf("Got %v response to get-door-delay request for %v", delay, deviceID)
		} else if door, ok := c.Door(delay.Door); ok {
			catalog.PutV(door, DoorDelay, delay.Delay)
		}
	}

	for _, d := range []uint8{1, 2, 3, 4} {
		if control, err := api.GetDoorControl(uhppoted.GetDoorControlRequest{DeviceID: deviceID, Door: d}); err != nil {
			log.Warnf("%v", err)
		} else if control == nil {
			log.Warnf("Got %v response to get-door-control request for %v", control, deviceID)
		} else if door, ok := c.Door(control.Door); ok {
			catalog.PutV(door, DoorControl, control.Control)
		}
	}
}

func (l *LAN) GetEvents(c Controller, intervals []types.Interval) {
	api := l.api([]Controller{c})
	deviceID := c.ID()
	oid := c.OIDx()

	log.Infof("%v: retrieving LAN controller events (%v)", deviceID, intervals)

	if first, last, current, err := api.GetEventIndices(deviceID); err != nil {
		log.Warnf("%v", err)
	} else {
		catalog.PutV(oid, ControllerTouched, time.Now())
		catalog.PutV(oid, ControllerEventsStatus, types.StatusOk)
		catalog.PutV(oid, ControllerEventsFirst, first)
		catalog.PutV(oid, ControllerEventsLast, last)
		catalog.PutV(oid, ControllerEventsCurrent, current)

		count := 0
		events := []uhppoted.Event{}

		f := func(index uint32) {
			log.Sayf("get event %v", index)
			if e, err := api.GetEvent(deviceID, index); err != nil {
				log.Warnf("%v", err)
			} else if e != nil {
				events = append(events, *e)
			}

			count++
		}

		for _, interval := range intervals {
			if interval.Contains(last) {
				index := interval.From
				for index <= last && count < MAX {
					f(index)
					index++
				}
			}

			if interval.Contains(first) {
				index := interval.To
				for index >= first && count < MAX {
					f(index)
					index--
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

func (l *LAN) SynchTime(c Controller) {
	api := l.api([]Controller{c})
	deviceID := uhppoted.DeviceID(c.ID())
	location := c.TimeZone()
	now := time.Now().In(location)
	datetime := core.DateTime(now)

	request := uhppoted.SetTimeRequest{
		DeviceID: deviceID,
		DateTime: datetime,
	}

	if response, err := api.SetTime(request); err != nil {
		log.Warnf("%v", err)
	} else if response != nil {
		catalog.PutV(c.OIDx(), ControllerDateTimeModified, false)

		if status, err := api.GetStatus(uhppoted.GetStatusRequest{DeviceID: deviceID}); err != nil {
			log.Warnf("%v", err)
		} else if status == nil {
			log.Warnf("Got %v response to get-status request for %v", status, deviceID)
		} else {
			catalog.PutV(c.OIDx(), ControllerDateTimeCurrent, status.Status.SystemDateTime)
		}

		log.Infof("synchronized device-time %v %v", response.DeviceID, response.DateTime)
	}
}

func (l *LAN) SynchDoors(c Controller) {
	api := l.api([]Controller{c})
	deviceID := c.ID()

	// ... update door delays
	for _, door := range []uint8{1, 2, 3, 4} {
		if oid, ok := c.Door(door); ok && oid != "" {
			configured := catalog.GetV(oid, DoorDelayConfigured)
			actual := catalog.GetV(oid, DoorDelay)
			modified := false

			if b, ok := catalog.GetBool(oid, schema.DoorDelayModified); ok {
				modified = b
			}

			if configured != nil && (actual == nil || actual != configured) && modified {
				delay := configured.(uint8)

				request := uhppoted.SetDoorDelayRequest{
					DeviceID: uhppoted.DeviceID(deviceID),
					Door:     door,
					Delay:    delay,
				}

				if response, err := api.SetDoorDelay(request); err != nil {
					log.Warnf("%v", err)
				} else if response != nil {
					catalog.PutV(oid, DoorDelay, delay)
					catalog.PutV(oid, DoorDelayModified, false)
					log.Infof("%v: synchronized door %v delay (%v)", response.DeviceID, door, delay)
				}
			}
		}
	}

	// ... update door control states
	for _, door := range []uint8{1, 2, 3, 4} {
		if oid, ok := c.Door(door); ok && oid != "" {
			configured := catalog.GetV(oid, DoorControlConfigured)
			actual := catalog.GetV(oid, DoorControl)
			modified := false

			if b, ok := catalog.GetBool(oid, DoorControlModified); ok {
				modified = b
			}

			if configured != nil && (actual == nil || actual != configured) && modified {
				mode := configured.(core.ControlState)

				request := uhppoted.SetDoorControlRequest{
					DeviceID: uhppoted.DeviceID(deviceID),
					Door:     door,
					Control:  mode,
				}

				if response, err := api.SetDoorControl(request); err != nil {
					log.Warnf("%v", err)
				} else if response != nil {
					catalog.PutV(oid, DoorControl, mode)
					catalog.PutV(oid, DoorControlModified, false)
					log.Infof("%v: synchronized door %v control (%v)", response.DeviceID, door, mode)
				}
			}
		}
	}
}

func (l *LAN) CompareACL(controllers []Controller, permissions acl.ACL) error {
	log.Debugf("%v", "Comparing ACL")

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

	compare, err := acl.Compare(permissions, current)
	if err != nil {
		return err
	} else if compare == nil {
		return fmt.Errorf("Invalid ACL compare report: %v", compare)
	}

	for k, v := range compare {
		log.Infof("ACL %v  unchanged:%-3v updated:%-3v added:%-3v deleted:%-3v", k, len(v.Unchanged), len(v.Updated), len(v.Added), len(v.Deleted))
	}

	diff := acl.SystemDiff(compare)
	report := diff.Consolidate()
	if report == nil {
		return fmt.Errorf("Invalid consolidated ACL compare report: %v", report)
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
					catalog.PutV(c.OIDx(), ControllerCardsStatus, types.StatusError)
				} else {
					catalog.PutV(c.OIDx(), ControllerCardsStatus, types.StatusOk)
				}
				break
			}
		}
	}

	return nil
}

func (l *LAN) UpdateACL(controllers []Controller, permissions acl.ACL) error {
	log.Infof("%v", "Updating ACL")

	api := l.api(controllers)
	rpt, errors := acl.PutACL(api.UHPPOTE, permissions, false)
	for _, err := range errors {
		log.Warnf("%v", err)
	}

	keys := []uint32{}
	for k, _ := range rpt {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	log.Infof("%v", "ACL updated")

	return nil
}

func (l *LAN) status() types.Status {
	return types.StatusOk
}

func (l *LAN) api(controllers []Controller) *uhppoted.UHPPOTED {
	devices := []uhppote.Device{}

	for _, v := range controllers {
		name := v.Name()
		id := v.ID()
		addr := v.EndPoint()

		if id > 0 && addr != nil {
			devices = append(devices, uhppote.Device{
				Name:     name,
				DeviceID: id,
				Address:  addr,
				Doors:    []string{},
				TimeZone: time.Local,
			})
		}
	}

	u := uhppote.NewUHPPOTE(l.BindAddress, l.BroadcastAddress, l.ListenAddress, 1*time.Second, devices, l.Debug)
	api := uhppoted.UHPPOTED{
		UHPPOTE: u,
		Log:     log.Default(),
	}

	return &api
}

func (l LAN) serialize() ([]byte, error) {
	record := struct {
		OID              schema.OID         `json:"OID"`
		Name             string             `json:"name,omitempty"`
		BindAddress      core.BindAddr      `json:"bind-address,omitempty"`
		BroadcastAddress core.BroadcastAddr `json:"broadcast-address,omitempty"`
		ListenAddress    core.ListenAddr    `json:"listen-address,omitempty"`
		Created          types.Timestamp    `json:"created,omitempty"`
		Modified         types.Timestamp    `json:"modified,omitempty"`
	}{
		OID:              l.OID,
		Name:             l.Name,
		BindAddress:      l.BindAddress,
		BroadcastAddress: l.BroadcastAddress,
		ListenAddress:    l.ListenAddress,
		Created:          l.created,
		Modified:         l.modified,
	}

	return json.MarshalIndent(record, "", "  ")
}

func (l *LAN) deserialize(bytes []byte) error {
	created = created.Add(1 * time.Minute)

	record := struct {
		OID              schema.OID         `json:"OID"`
		Name             string             `json:"name,omitempty"`
		BindAddress      core.BindAddr      `json:"bind-address,omitempty"`
		BroadcastAddress core.BroadcastAddr `json:"broadcast-address,omitempty"`
		ListenAddress    core.ListenAddr    `json:"listen-address,omitempty"`
		Created          types.Timestamp    `json:"created,omitempty"`
		Modified         types.Timestamp    `json:"modified,omitempty"`
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

func (l *LAN) log(uid string, operation string, OID schema.OID, field, description, before, after string, dbc db.DBC) {
	record := audit.AuditRecord{
		UID:       uid,
		OID:       OID,
		Component: "LAN",
		Operation: operation,
		Details: audit.Details{
			ID:          "LAN",
			Name:        stringify(l.Name, ""),
			Field:       field,
			Description: description,
			Before:      before,
			After:       after,
		},
	}

	if dbc != nil {
		dbc.Write(record)
	}
}
