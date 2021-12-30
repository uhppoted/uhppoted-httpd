package interfaces

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"sort"
	"strings"
	"time"

	core "github.com/uhppoted/uhppote-core/types"
	"github.com/uhppoted/uhppote-core/uhppote"
	"github.com/uhppoted/uhppoted-lib/acl"
	"github.com/uhppoted/uhppoted-lib/uhppoted"

	"github.com/uhppoted/uhppoted-httpd/audit"
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/db"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type LAN struct {
	OID              catalog.OID
	Name             string
	BindAddress      core.BindAddr
	BroadcastAddress core.BroadcastAddr
	ListenAddress    core.ListenAddr
	Debug            bool

	ch           chan types.EventsList
	created      types.DateTime
	deleted      *types.DateTime
	unconfigured bool
}

type Controller interface {
	OID() catalog.OID
	Name() string
	DeviceID() uint32
	EndPoint() *net.UDPAddr
	TimeZone() *time.Location
	Door(uint8) (catalog.OID, bool)
}

var created = types.DateTimeNow()

func (l LAN) String() string {
	return fmt.Sprintf("%v", l.Name)
}

func (l *LAN) IsValid() bool {
	if l != nil {
		if strings.TrimSpace(l.Name) != "" {
			return true
		}
	}

	return false
}

func (l *LAN) IsDeleted() bool {
	if l != nil && l.deleted != nil {
		return true
	}

	return false
}

func (l *LAN) AsObjects(a auth.OpAuth) []interface{} {
	type E = struct {
		field catalog.Suffix
		value interface{}
	}

	list := []E{}

	if l.deleted != nil {
		list = append(list, E{LANDeleted, l.deleted})
	} else {
		list = append(list, E{LANStatus, l.status()})
		list = append(list, E{LANCreated, l.created})
		list = append(list, E{LANDeleted, l.deleted})
		list = append(list, E{LANType, "LAN"})
		list = append(list, E{LANName, l.Name})
		list = append(list, E{LANBindAddress, l.BindAddress})
		list = append(list, E{LANBroadcastAddress, l.BroadcastAddress})
		list = append(list, E{LANListenAddress, l.ListenAddress})
	}

	f := func(l *LAN, field string, value interface{}) bool {
		if a != nil {
			if err := a.CanView(auth.Interfaces, l, field, value); err != nil {
				return false
			}
		}

		return true
	}

	objects := []interface{}{}

	if l.deleted == nil && f(l, "OID", l.OID) {
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

func (l *LAN) AsRuleEntity() interface{} {
	type entity struct {
		Type string
		Name string
	}

	if l != nil {
		return &entity{
			Type: "LAN",
			Name: fmt.Sprintf("%v", l.Name),
		}
	}

	return &entity{}
}

func (l *LAN) set(auth auth.OpAuth, oid catalog.OID, value string, dbc db.DBC) ([]catalog.Object, error) {
	objects := []catalog.Object{}

	f := func(field string, value interface{}) error {
		if auth == nil {
			return nil
		}

		return auth.CanUpdateInterface(l, field, value)
	}

	if l == nil {
		return objects, nil
	} else if l.deleted != nil {
		objects = append(objects, catalog.NewObject2(l.OID, LANDeleted, l.deleted))

		return objects, fmt.Errorf("LAN has been deleted")
	}

	switch oid {
	case l.OID.Append(LANName):
		if err := f("name", value); err != nil {
			return nil, err
		} else {
			l.log(auth,
				"update",
				l.OID,
				"name",
				fmt.Sprintf("Updated name from %v to %v", stringify(l.Name, BLANK), stringify(value, BLANK)),
				stringify(l.Name, BLANK),
				stringify(value, BLANK),
				dbc)
			l.Name = value
			objects = append(objects, catalog.NewObject2(l.OID, LANName, l.Name))
		}

	case l.OID.Append(LANBindAddress):
		if l.deleted != nil {
			return nil, fmt.Errorf("LAN has been deleted")
		} else if addr, err := core.ResolveBindAddr(value); err != nil {
			return nil, err
		} else if err := f("bind", addr); err != nil {
			return nil, err
		} else {
			l.log(auth,
				"update",
				l.OID,
				"bind",
				fmt.Sprintf("Updated bind address from %v to %v", stringify(l.BindAddress, BLANK), stringify(value, BLANK)),
				stringify(l.BindAddress, BLANK),
				stringify(value, BLANK),
				dbc)
			l.BindAddress = *addr
			objects = append(objects, catalog.NewObject2(l.OID, LANBindAddress, l.BindAddress))
		}

	case l.OID.Append(LANBroadcastAddress):
		if l.deleted != nil {
			return nil, fmt.Errorf("LAN has been deleted")
		} else if addr, err := core.ResolveBroadcastAddr(value); err != nil {
			return nil, err
		} else if err := f("broadcast", addr); err != nil {
			return nil, err
		} else {
			l.log(auth,
				"update",
				l.OID,
				"broadcast",
				fmt.Sprintf("Updated broadcast address from %v to %v", stringify(l.BroadcastAddress, BLANK), stringify(value, BLANK)),
				stringify(l.BroadcastAddress, BLANK),
				stringify(value, BLANK),
				dbc)
			l.BroadcastAddress = *addr
			objects = append(objects, catalog.NewObject2(l.OID, LANBroadcastAddress, l.BroadcastAddress))
		}

	case l.OID.Append(LANListenAddress):
		if l.deleted != nil {
			return nil, fmt.Errorf("LAN has been deleted")
		} else if addr, err := core.ResolveListenAddr(value); err != nil {
			return nil, err
		} else if err = f("listen", addr); err != nil {
			return nil, err
		} else {
			l.log(auth,
				"update",
				l.OID,
				"listen",
				fmt.Sprintf("Updated listen address from %v to %v", stringify(l.ListenAddress, BLANK), stringify(value, BLANK)),
				stringify(l.ListenAddress, BLANK),
				stringify(value, BLANK),
				dbc)
			l.ListenAddress = *addr
			objects = append(objects, catalog.NewObject2(l.OID, LANListenAddress, l.ListenAddress))
		}
	}

	if l.deleted == nil {
		objects = append(objects, catalog.NewObject2(l.OID, LANStatus, l.status()))
	}

	return objects, nil
}

func (l *LAN) status() types.Status {
	return types.StatusOk
}

func (l LAN) Clone() LAN {
	return LAN{
		OID:              l.OID,
		Name:             l.Name,
		BindAddress:      l.BindAddress,
		BroadcastAddress: l.BroadcastAddress,
		ListenAddress:    l.ListenAddress,
		Debug:            l.Debug,

		created:      l.created,
		deleted:      l.deleted,
		unconfigured: l.unconfigured,
	}
}

func (l *LAN) api(controllers []Controller) *uhppoted.UHPPOTED {
	devices := []uhppote.Device{}

	for _, v := range controllers {
		name := v.Name()
		id := v.DeviceID()
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
		Log:     log.New(os.Stdout, "", log.LstdFlags|log.LUTC),
	}

	return &api
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
	log.Printf("%v: refreshing LAN controller status", c.DeviceID())

	api := l.api([]Controller{c})
	deviceID := uhppoted.DeviceID(c.DeviceID())

	if info, err := api.GetDevice(uhppoted.GetDeviceRequest{DeviceID: deviceID}); err != nil {
		log.Printf("%v", err)
	} else if info == nil {
		log.Printf("Got %v response to get-device request for %v", info, deviceID)
	} else {
		l.store(c, *info)
	}

	if status, err := api.GetStatus(uhppoted.GetStatusRequest{DeviceID: deviceID}); err != nil {
		log.Printf("%v", err)
	} else if status == nil {
		log.Printf("Got %v response to get-status request for %v", status, deviceID)
	} else {
		l.store(c, *status)
	}

	if cards, err := api.GetCardRecords(uhppoted.GetCardRecordsRequest{DeviceID: deviceID}); err != nil {
		log.Printf("%v", err)
	} else if cards == nil {
		log.Printf("Got %v response to get-card-records request for %v", cards, deviceID)
	} else {
		l.store(c, *cards)
	}

	if events, err := api.GetEventRange(uhppoted.GetEventRangeRequest{DeviceID: deviceID}); err != nil {
		log.Printf("%v", err)
	} else if events == nil {
		log.Printf("Got %v response to get-event-range request for %v", events, deviceID)
	} else {
		l.store(c, *events)
	}

	for _, d := range []uint8{1, 2, 3, 4} {
		if delay, err := api.GetDoorDelay(uhppoted.GetDoorDelayRequest{DeviceID: deviceID, Door: d}); err != nil {
			log.Printf("%v", err)
		} else if delay == nil {
			log.Printf("Got %v response to get-door-delay request for %v", delay, deviceID)
		} else {
			l.store(c, *delay)
		}
	}

	for _, d := range []uint8{1, 2, 3, 4} {
		if control, err := api.GetDoorControl(uhppoted.GetDoorControlRequest{DeviceID: deviceID, Door: d}); err != nil {
			log.Printf("%v", err)
		} else if control == nil {
			log.Printf("Got %v response to get-door-control request for %v", control, deviceID)
		} else {
			l.store(c, *control)
		}
	}

	if recent, err := api.GetEvents(uhppoted.GetEventsRequest{DeviceID: deviceID, Max: 5}); err != nil {
		log.Printf("%v", err)
	} else if l.ch != nil {
		l.ch <- types.EventsList{
			DeviceID: uint32(recent.DeviceID),
			Events:   recent.Events,
		}
	}
}

func (l *LAN) SynchTime(c Controller) {
	api := l.api([]Controller{c})
	deviceID := c.DeviceID()
	location := c.TimeZone()
	now := time.Now().In(location)
	datetime := core.DateTime(now)

	request := uhppoted.SetTimeRequest{
		DeviceID: uhppoted.DeviceID(deviceID),
		DateTime: datetime,
	}

	if response, err := api.SetTime(request); err != nil {
		log.Printf("ERROR %v", err)
	} else if response != nil {
		log.Printf("INFO  synchronized device-time %v %v", response.DeviceID, response.DateTime)
	}
}

func (l *LAN) SynchDoors(c Controller) {
	api := l.api([]Controller{c})
	deviceID := c.DeviceID()

	// ... update door delays
	for _, door := range []uint8{1, 2, 3, 4} {
		if oid, ok := c.Door(door); ok && oid != "" {
			configured := catalog.GetV(oid, catalog.DoorDelayConfigured)
			actual := catalog.GetV(oid, catalog.DoorDelay)
			modified := false

			if v := catalog.GetV(oid, catalog.DoorDelayModified); v != nil {
				if b, ok := v.(bool); ok {
					modified = b
				}
			}

			if configured != nil && (actual == nil || actual != configured) && modified {
				delay := configured.(uint8)

				request := uhppoted.SetDoorDelayRequest{
					DeviceID: uhppoted.DeviceID(deviceID),
					Door:     door,
					Delay:    delay,
				}

				if response, err := api.SetDoorDelay(request); err != nil {
					warn(err)
				} else if response != nil {
					catalog.PutV(oid, DoorDelay, delay)
					catalog.PutV(oid, DoorDelayModified, false)
					info(fmt.Sprintf("%v: synchronized door %v delay (%v)", response.DeviceID, door, delay))
				}
			}
		}
	}

	// ... update door control states
	for _, door := range []uint8{1, 2, 3, 4} {
		if oid, ok := c.Door(door); ok && oid != "" {
			configured := catalog.GetV(oid, catalog.DoorControlConfigured)
			actual := catalog.GetV(oid, catalog.DoorControl)
			modified := false

			if v := catalog.GetV(oid, catalog.DoorControlModified); v != nil {
				if b, ok := v.(bool); ok {
					modified = b
				}
			}

			if configured != nil && (actual == nil || actual != configured) && modified {
				mode := configured.(core.ControlState)

				request := uhppoted.SetDoorControlRequest{
					DeviceID: uhppoted.DeviceID(deviceID),
					Door:     door,
					Control:  mode,
				}

				if response, err := api.SetDoorControl(request); err != nil {
					warn(err)
				} else if response != nil {
					catalog.PutV(oid, DoorControl, mode)
					catalog.PutV(oid, DoorControlModified, false)
					info(fmt.Sprintf("%v: synchronized door %v control (%v)", response.DeviceID, door, mode))
				}
			}
		}
	}
}

func (l *LAN) CompareACL(controllers []Controller, permissions acl.ACL) error {
	log.Printf("Comparing ACL")

	devices := []uhppote.Device{}
	api := l.api(controllers)
	for _, v := range api.UHPPOTE.DeviceList() {
		device := v
		devices = append(devices, device)
	}

	current, errors := acl.GetACL(api.UHPPOTE, devices)
	for _, err := range errors {
		warn(err)
	}

	compare, err := acl.Compare(permissions, current)
	if err != nil {
		return err
	} else if compare == nil {
		return fmt.Errorf("Invalid ACL compare report: %v", compare)
	}

	for k, v := range compare {
		log.Printf("ACL %v - unchanged:%-3v updated:%-3v added:%-3v deleted:%-3v", k, len(v.Unchanged), len(v.Updated), len(v.Added), len(v.Deleted))
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

	log.Printf("ACL compare - unchanged:%-3v updated:%-3v added:%-3v deleted:%-3v", unchanged, updated, added, deleted)

	for _, c := range controllers {
		for _, d := range devices {
			if c.DeviceID() == d.DeviceID {
				rs := compare[c.DeviceID()]
				l.store(c, rs)
				break
			}
		}
	}

	return nil
}

func (l *LAN) UpdateACL(controllers []Controller, permissions acl.ACL) error {
	log.Printf("Updating ACL")

	api := l.api(controllers)
	rpt, errors := acl.PutACL(api.UHPPOTE, permissions, false)
	for _, err := range errors {
		warn(err)
	}

	keys := []uint32{}
	for k, _ := range rpt {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	var msg bytes.Buffer
	fmt.Fprintf(&msg, "ACL updated\n")

	for _, k := range keys {
		v := rpt[k]
		fmt.Fprintf(&msg, "                    %v", k)
		fmt.Fprintf(&msg, " unchanged:%-3v", len(v.Unchanged))
		fmt.Fprintf(&msg, " updated:%-3v", len(v.Updated))
		fmt.Fprintf(&msg, " added:%-3v", len(v.Added))
		fmt.Fprintf(&msg, " deleted:%-3v", len(v.Deleted))
		fmt.Fprintf(&msg, " failed:%-3v", len(v.Failed))
		fmt.Fprintf(&msg, " errored:%-3v", len(v.Errored))
		fmt.Fprintln(&msg)
	}

	log.Printf("%v", string(msg.Bytes()))

	return nil
}

func (l *LAN) store(c Controller, info interface{}) {
	switch v := info.(type) {
	case uhppoted.GetDeviceResponse:
		addr := core.Address(v.Address)
		catalog.PutV(c.OID(), ControllerTouched, time.Now())
		catalog.PutV(c.OID(), ControllerEndpointAddress, addr)

	case uhppoted.GetStatusResponse:
		datetime := types.DateTime(v.Status.SystemDateTime)
		catalog.PutV(c.OID(), ControllerTouched, time.Now())
		catalog.PutV(c.OID(), ControllerDateTimeCurrent, datetime)

	case uhppoted.GetCardRecordsResponse:
		cards := v.Cards
		catalog.PutV(c.OID(), ControllerTouched, time.Now())
		catalog.PutV(c.OID(), ControllerCardsCount, cards)

	case uhppoted.GetEventRangeResponse:
		events := v.Events.Last
		catalog.PutV(c.OID(), ControllerTouched, time.Now())
		catalog.PutV(c.OID(), ControllerEventsCount, events)

	case uhppoted.GetDoorDelayResponse:
		if door, ok := c.Door(v.Door); ok {
			catalog.PutV(door, DoorDelay, v.Delay)
		}

	case uhppoted.GetDoorControlResponse:
		if door, ok := c.Door(v.Door); ok {
			catalog.PutV(door, DoorControl, v.Control)
		}

	case acl.Diff:
		if len(v.Updated)+len(v.Added)+len(v.Deleted) > 0 {
			catalog.PutV(c.OID(), ControllerCardsStatus, types.StatusError)
		} else {
			catalog.PutV(c.OID(), ControllerCardsStatus, types.StatusOk)
		}
	}
}
func (l LAN) serialize() ([]byte, error) {
	record := struct {
		OID              catalog.OID        `json:"OID"`
		Name             string             `json:"name,omitempty"`
		BindAddress      core.BindAddr      `json:"bind-address,omitempty"`
		BroadcastAddress core.BroadcastAddr `json:"broadcast-address,omitempty"`
		ListenAddress    core.ListenAddr    `json:"listen-address,omitempty"`
		Created          types.DateTime     `json:"created,omitempty"`
	}{
		OID:              l.OID,
		Name:             l.Name,
		BindAddress:      l.BindAddress,
		BroadcastAddress: l.BroadcastAddress,
		ListenAddress:    l.ListenAddress,
		Created:          types.DateTime(l.created),
	}

	return json.MarshalIndent(record, "", "  ")
}

func (l *LAN) deserialize(bytes []byte) error {
	created = created.Add(1 * time.Minute)

	record := struct {
		OID              catalog.OID        `json:"OID"`
		Name             string             `json:"name,omitempty"`
		BindAddress      core.BindAddr      `json:"bind-address,omitempty"`
		BroadcastAddress core.BroadcastAddr `json:"broadcast-address,omitempty"`
		ListenAddress    core.ListenAddr    `json:"listen-address,omitempty"`
		Created          types.DateTime     `json:"created,omitempty"`
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
	l.unconfigured = false

	return nil
}

func (l *LAN) log(auth auth.OpAuth, operation string, OID catalog.OID, field, description, before, after string, dbc db.DBC) {
	uid := ""
	if auth != nil {
		uid = auth.UID()
	}

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
