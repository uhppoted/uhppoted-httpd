package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"sort"
	"sync"
	"time"

	core "github.com/uhppoted/uhppote-core/types"
	"github.com/uhppoted/uhppote-core/uhppote"
	"github.com/uhppoted/uhppoted-httpd/audit"
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/db"
	"github.com/uhppoted/uhppoted-httpd/types"
	"github.com/uhppoted/uhppoted-lib/acl"
	"github.com/uhppoted/uhppoted-lib/uhppoted"
)

type LAN struct {
	OID              catalog.OID        `json:"OID"`
	Name             string             `json:"name"`
	BindAddress      core.BindAddr      `json:"bind-address"`
	BroadcastAddress core.BroadcastAddr `json:"broadcast-address"`
	ListenAddress    core.ListenAddr    `json:"listen-address"`
	Debug            bool               `json:"debug"`

	status       types.Status
	created      time.Time
	deleted      *time.Time
	unconfigured bool
}

type device struct {
	touched  time.Time
	address  *core.Address
	datetime *types.DateTime
	cards    *uint32
	events   *uint32
	acl      types.Status
}

var cache = struct {
	cache map[uint32]device
	guard sync.RWMutex
}{
	cache: map[uint32]device{},
}

func (l *LAN) String() string {
	return fmt.Sprintf("%v", l.Name)
}

func (l *LAN) AsObjects() []interface{} {
	objects := []interface{}{
		catalog.NewObject(l.OID, types.StatusUncertain),
		catalog.NewObject2(l.OID, Status, l.status),
		catalog.NewObject2(l.OID, LANType, "LAN"),
		catalog.NewObject2(l.OID, LANName, l.Name),
		catalog.NewObject2(l.OID, LANBindAddress, l.BindAddress),
		catalog.NewObject2(l.OID, LANBroadcastAddress, l.BroadcastAddress),
		catalog.NewObject2(l.OID, LANListenAddress, l.ListenAddress),
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

func (l *LAN) clone() *LAN {
	if l != nil {
		lan := LAN{
			OID:              l.OID,
			Name:             l.Name,
			BindAddress:      l.BindAddress,
			BroadcastAddress: l.BroadcastAddress,
			ListenAddress:    l.ListenAddress,
			Debug:            l.Debug,

			status:       l.status,
			created:      l.created,
			deleted:      l.deleted,
			unconfigured: l.unconfigured,
		}

		return &lan
	}

	return nil
}

func (l *LAN) set(auth auth.OpAuth, oid catalog.OID, value string, dbc db.DBC) ([]catalog.Object, error) {
	objects := []catalog.Object{}

	f := func(field string, value interface{}) error {
		if auth == nil {
			return nil
		}

		return auth.CanUpdateInterface(l, field, value)
	}

	if l != nil {
		switch oid {
		case l.OID.Append(LANName):
			if err := f("name", value); err != nil {
				return nil, err
			} else {
				l.log(auth, "update", l.OID, "name", fmt.Sprintf("Updated name from %v to %v", stringify(l.Name, "<blank>"), stringify(value, "<blank>")), dbc)
				l.Name = value
				objects = append(objects, catalog.NewObject2(l.OID, LANName, l.Name))
			}

		case l.OID.Append(LANBindAddress):
			if addr, err := core.ResolveBindAddr(value); err != nil {
				return nil, err
			} else if err := f("bind", addr); err != nil {
				return nil, err
			} else {
				l.log(auth, "update", l.OID, "bind", fmt.Sprintf("Updated bind address from %v to %v", stringify(l.BindAddress, "<blank>"), stringify(value, "<blank>")), dbc)
				l.BindAddress = *addr
				objects = append(objects, catalog.NewObject2(l.OID, LANBindAddress, l.BindAddress))
			}

		case l.OID.Append(LANBroadcastAddress):
			if addr, err := core.ResolveBroadcastAddr(value); err != nil {
				return nil, err
			} else if err := f("broadcast", addr); err != nil {
				return nil, err
			} else {
				l.log(auth, "update", l.OID, "broadcast", fmt.Sprintf("Updated broadcast address from %v to %v", stringify(l.BroadcastAddress, "<blank>"), stringify(value, "<blank>")), dbc)
				l.BroadcastAddress = *addr
				objects = append(objects, catalog.NewObject2(l.OID, LANBroadcastAddress, l.BroadcastAddress))
			}

		case l.OID.Append(LANListenAddress):
			if addr, err := core.ResolveListenAddr(value); err != nil {
				return nil, err
			} else if err = f("listen", addr); err != nil {
				return nil, err
			} else {
				l.log(auth, "update", l.OID, "listen", fmt.Sprintf("Updated listen address from %v to %v", stringify(l.ListenAddress, "<blank>"), stringify(value, "<blank>")), dbc)
				l.ListenAddress = *addr
				objects = append(objects, catalog.NewObject2(l.OID, LANListenAddress, l.ListenAddress))
			}
		}
	}

	return objects, nil
}

func (l *LAN) api(controllers []*Controller) *uhppoted.UHPPOTED {
	devices := []uhppote.Device{}

	for _, v := range controllers {
		if v.DeviceID == nil || *v.DeviceID == 0 || v.IP == nil {
			continue
		}

		name := v.Name.String()
		id := *v.DeviceID
		addr := net.UDPAddr(*v.IP)

		devices = append(devices, uhppote.Device{
			Name:     name,
			DeviceID: id,
			Address:  &addr,
			Rollover: 100000,
			Doors:    []string{},
			TimeZone: time.Local,
		})
	}

	u := uhppote.NewUHPPOTE(l.BindAddress, l.BroadcastAddress, l.ListenAddress, 1*time.Second, devices, l.Debug)
	api := uhppoted.UHPPOTED{
		UHPPOTE: u,
		Log:     log.New(os.Stdout, "", log.LstdFlags|log.LUTC),
	}

	return &api
}

func (l *LAN) updateACL(controllers []*Controller, permissions acl.ACL) {
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
}

func (l *LAN) compareACL(controllers []*Controller, permissions acl.ACL) error {
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

	for _, v := range devices {
		id := v.DeviceID
		l.store(id, compare[id], nil)
	}

	return nil
}

// Possibly a long-running function - expects to be invoked from an external goroutine
func (l *LAN) refresh(controllers []*Controller, callback Callback) []catalog.Object {
	var lock sync.Mutex
	var objects = []catalog.Object{}

	expired := time.Now().Add(-windows.cacheExpiry)
	for k, v := range cache.cache {
		if v.touched.Before(expired) {
			delete(cache.cache, k)
			log.Printf("Controller %v cache entry expired", k)
		}
	}

	list := map[uint32]struct{}{}
	for _, c := range controllers {
		if c.DeviceID != nil && *c.DeviceID != 0 {
			list[*c.DeviceID] = struct{}{}
		}
	}

	api := l.api(controllers)
	if devices, err := api.GetDevices(uhppoted.GetDevicesRequest{}); err != nil {
		log.Printf("%v", err)
	} else if devices == nil {
		log.Printf("Got %v response to get-devices request", devices)
	} else {
		for k, v := range devices.Devices {
			if d, ok := api.UHPPOTE.DeviceList()[k]; ok {
				d.Address.IP = v.Address
				d.Address.Port = v.Port
			}

			list[k] = struct{}{}
		}
	}

	var wg sync.WaitGroup

	for k, _ := range list {
		wg.Add(1)

		id := k
		go func() {
			defer wg.Done()

			var controller *Controller
			for _, c := range controllers {
				if c.DeviceID != nil && *c.DeviceID == id {
					controller = c
					break
				}
			}

			if list := l.update(api, id, controller, callback); list != nil {
				lock.Lock()
				defer lock.Unlock()
				objects = append(objects, list...)
			}
		}()
	}

	wg.Wait()

	return objects
}

func (l *LAN) update(api *uhppoted.UHPPOTED, id uint32, controller *Controller, callback Callback) []catalog.Object {
	objects := []catalog.Object{}

	log.Printf("%v: refreshing LAN controller status", id)

	if info, err := api.GetDevice(uhppoted.GetDeviceRequest{DeviceID: uhppoted.DeviceID(id)}); err != nil {
		log.Printf("%v", err)
	} else if info == nil {
		log.Printf("Got %v response to get-device request for %v", info, id)
	} else if list := l.store(id, *info, controller); list != nil {
		objects = append(objects, list...)
	}

	if status, err := api.GetStatus(uhppoted.GetStatusRequest{DeviceID: uhppoted.DeviceID(id)}); err != nil {
		log.Printf("%v", err)
	} else if status == nil {
		log.Printf("Got %v response to get-status request for %v", status, id)
	} else if list := l.store(id, *status, controller); list != nil {
		objects = append(objects, list...)
	}

	if cards, err := api.GetCardRecords(uhppoted.GetCardRecordsRequest{DeviceID: uhppoted.DeviceID(id)}); err != nil {
		log.Printf("%v", err)
	} else if cards == nil {
		log.Printf("Got %v response to get-card-records request for %v", cards, id)
	} else if list := l.store(id, *cards, controller); list != nil {
		objects = append(objects, list...)
	}

	if events, err := api.GetEventRange(uhppoted.GetEventRangeRequest{DeviceID: uhppoted.DeviceID(id)}); err != nil {
		log.Printf("%v", err)
	} else if events == nil {
		log.Printf("Got %v response to get-event-range request for %v", events, id)
	} else if list := l.store(id, *events, controller); list != nil {
		objects = append(objects, list...)
	}

	for _, d := range []uint8{1, 2, 3, 4} {
		if delay, err := api.GetDoorDelay(uhppoted.GetDoorDelayRequest{DeviceID: uhppoted.DeviceID(id), Door: d}); err != nil {
			log.Printf("%v", err)
		} else if delay == nil {
			log.Printf("Got %v response to get-door-delay request for %v", delay, id)
		} else if list := l.store(id, *delay, controller); list != nil {
			objects = append(objects, list...)
		}
	}

	for _, d := range []uint8{1, 2, 3, 4} {
		if control, err := api.GetDoorControl(uhppoted.GetDoorControlRequest{DeviceID: uhppoted.DeviceID(id), Door: d}); err != nil {
			log.Printf("%v", err)
		} else if control == nil {
			log.Printf("Got %v response to get-door-control request for %v", control, id)
		} else if list := l.store(id, *control, controller); list != nil {
			objects = append(objects, list...)
		}
	}

	if recent, err := api.GetEvents(uhppoted.GetEventsRequest{DeviceID: uhppoted.DeviceID(id), Max: 5}); err != nil {
		log.Printf("%v", err)
	} else if callback != nil {
		callback.Append(id, recent.Events)
	}

	return objects
}

func (l *LAN) store(id uint32, info interface{}, controller *Controller) []catalog.Object {
	cache.guard.Lock()
	defer cache.guard.Unlock()

	objects := []catalog.Object{}
	cached, ok := cache.cache[id]
	if !ok {
		cached = device{}
	}

	switch v := info.(type) {
	case uhppoted.GetDeviceResponse:
		addr := core.Address(v.Address)
		cached.address = &addr
		cached.touched = time.Now()
		cache.cache[id] = cached

	case uhppoted.GetStatusResponse:
		datetime := types.DateTime(v.Status.SystemDateTime)
		cached.datetime = &datetime
		cached.touched = time.Now()
		cache.cache[id] = cached

	case uhppoted.GetCardRecordsResponse:
		cards := v.Cards
		cached.cards = &cards
		cached.touched = time.Now()
		cache.cache[id] = cached

	case uhppoted.GetEventRangeResponse:
		events := v.Events.Last
		cached.events = events
		cached.touched = time.Now()
		cache.cache[id] = cached

	case uhppoted.GetDoorDelayResponse:
		if controller != nil {
			if door, ok := controller.Doors[v.Door]; ok {
				objects = append(objects, catalog.NewObject2(door, DoorDelay, v.Delay))
			}
		}

	case uhppoted.GetDoorControlResponse:
		if controller != nil {
			if door, ok := controller.Doors[v.Door]; ok {
				objects = append(objects, catalog.NewObject2(door, DoorControl, v.Control))
			}
		}

	case acl.Diff:
		if ok {
			if len(v.Updated)+len(v.Added)+len(v.Deleted) > 0 {
				cached.acl = types.StatusError
			} else {
				cached.acl = types.StatusOk
			}

			cache.cache[id] = cached
		}
	}

	return objects
}

func (l *LAN) synchTime(controllers []*Controller) []catalog.Object {
	objects := []catalog.Object{}
	api := l.api(controllers)
	for _, c := range controllers {
		if c.DeviceID != nil {
			device := uhppoted.DeviceID(*c.DeviceID)
			location := time.Local

			if c.TimeZone != nil {
				timezone := *c.TimeZone
				if tz, err := types.Timezone(timezone); err == nil && tz != nil {
					location = tz
				}
			}

			now := time.Now().In(location)
			datetime := core.DateTime(now)

			request := uhppoted.SetTimeRequest{
				DeviceID: device,
				DateTime: datetime,
			}

			if response, err := api.SetTime(request); err != nil {
				log.Printf("ERROR %v", err)
			} else if response != nil {
				log.Printf("INFO  synchronized device-time %v %v", response.DeviceID, response.DateTime)
			}
		}
	}

	return objects
}

func (l *LAN) synchDoors(controllers []*Controller) []catalog.Object {
	objects := []catalog.Object{}
	api := l.api(controllers)

	for _, c := range controllers {
		if c.DeviceID != nil {
			device := uhppoted.DeviceID(*c.DeviceID)

			// ... update door delays
			for _, door := range []uint8{1, 2, 3, 4} {
				if oid, ok := c.Doors[door]; ok && oid != "" {
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
							DeviceID: device,
							Door:     door,
							Delay:    delay,
						}

						if response, err := api.SetDoorDelay(request); err != nil {
							log.Printf("ERROR %v", err)
						} else if response != nil {
							objects = append(objects, catalog.NewObject2(oid, DoorDelay, delay))
							objects = append(objects, catalog.NewObject2(oid, DoorDelayModified, false))
							log.Printf("INFO  %v: synchronized door %v delay (%v)", response.DeviceID, door, delay)
						}
					}
				}
			}

			// ... update door control states
			for _, door := range []uint8{1, 2, 3, 4} {
				if oid, ok := c.Doors[door]; ok && oid != "" {
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
							DeviceID: device,
							Door:     door,
							Control:  mode,
						}

						if response, err := api.SetDoorControl(request); err != nil {
							log.Printf("ERROR %v", err)
						} else if response != nil {
							objects = append(objects, catalog.NewObject2(oid, DoorControl, mode))
							objects = append(objects, catalog.NewObject2(oid, DoorControlModified, false))
							log.Printf("INFO  %v: synchronized door %v control (%v)", response.DeviceID, door, mode)
						}
					}
				}
			}
		}
	}

	return objects
}

func (l *LAN) serialize() ([]byte, error) {
	if l == nil || l.deleted != nil || l.unconfigured {
		return nil, nil
	}

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
	created := types.DateTime(time.Now())
	record := struct {
		OID              catalog.OID        `json:"OID"`
		Name             string             `json:"name,omitempty"`
		BindAddress      core.BindAddr      `json:"bind-address,omitempty"`
		BroadcastAddress core.BroadcastAddr `json:"broadcast-address,omitempty"`
		ListenAddress    core.ListenAddr    `json:"listen-address,omitempty"`
		Created          *types.DateTime    `json:"created,omitempty"`
	}{
		Created: &created,
	}

	if err := json.Unmarshal(bytes, &record); err != nil {
		return err
	}

	l.OID = record.OID
	l.Name = record.Name
	l.BindAddress = record.BindAddress
	l.BroadcastAddress = record.BroadcastAddress
	l.ListenAddress = record.ListenAddress
	l.status = types.StatusOk
	l.created = time.Time(*record.Created)

	return nil
}

func (l *LAN) log(auth auth.OpAuth, operation string, OID catalog.OID, field string, description string, dbc db.DBC) {
	uid := ""
	if auth != nil {
		uid = auth.UID()
	}

	record := audit.AuditRecord{
		UID:       uid,
		OID:       OID,
		Component: "interface",
		Operation: operation,
		Details: audit.Details{
			ID:          "LAN",
			Name:        stringify(l.Name, ""),
			Field:       field,
			Description: description,
		},
	}

	if dbc != nil {
		dbc.Write(record)
	}
}
