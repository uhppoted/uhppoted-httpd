package controllers

import (
	"bytes"
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

	status types.Status
}

type device struct {
	touched  time.Time
	address  *core.Address
	datetime *types.DateTime
	cards    *uint32
	events   *uint32
	acl      types.Status
}

const LANName = catalog.InterfaceName
const LANType = catalog.InterfaceType
const LANBindAddress = catalog.LANBindAddress
const LANBroadcastAddress = catalog.LANBroadcastAddress
const LANListenAddress = catalog.LANListenAddress

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
		catalog.NewObject(l.OID, l.status),
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

			status: l.status,
		}

		return &lan
	}

	return nil
}

func (l *LAN) set(auth auth.OpAuth, oid catalog.OID, value string) ([]interface{}, error) {
	objects := []interface{}{}

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
				l.log(auth, "update", l.OID, "name", fmt.Sprintf("Updated name from %v to %v", stringify(l.Name, "<blank>"), stringify(value, "<blank>")))
				l.Name = value
				objects = append(objects, catalog.NewObject2(l.OID, LANName, l.Name))
			}

		case l.OID.Append(LANBindAddress):
			if addr, err := core.ResolveBindAddr(value); err != nil {
				return nil, err
			} else if err := f("bind", addr); err != nil {
				return nil, err
			} else {
				l.log(auth, "update", l.OID, "bind", fmt.Sprintf("Updated bind address from %v to %v", stringify(l.BindAddress, "<blank>"), stringify(value, "<blank>")))
				l.BindAddress = *addr
				objects = append(objects, catalog.NewObject2(l.OID, LANBindAddress, l.BindAddress))
			}

		case l.OID.Append(LANBroadcastAddress):
			if addr, err := core.ResolveBroadcastAddr(value); err != nil {
				return nil, err
			} else if err := f("broadcast", addr); err != nil {
				return nil, err
			} else {
				l.log(auth, "update", l.OID, "broadcast", fmt.Sprintf("Updated broadcast address from %v to %v", stringify(l.BroadcastAddress, "<blank>"), stringify(value, "<blank>")))
				l.BroadcastAddress = *addr
				objects = append(objects, catalog.NewObject2(l.OID, LANBroadcastAddress, l.BroadcastAddress))
			}

		case l.OID.Append(LANListenAddress):
			if addr, err := core.ResolveListenAddr(value); err != nil {
				return nil, err
			} else if err = f("listen", addr); err != nil {
				return nil, err
			} else {
				l.log(auth, "update", l.OID, "listen", fmt.Sprintf("Updated listen address from %v to %v", stringify(l.ListenAddress, "<blank>"), stringify(value, "<blank>")))
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

func (l *LAN) refresh(controllers []*Controller, callback Callback) {
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
	go func() {
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

		for k, _ := range list {
			id := k
			go func() {
				var controller *Controller
				for _, c := range controllers {
					if c.DeviceID != nil && *c.DeviceID == id {
						controller = c
						break
					}
				}

				l.update(api, id, controller, callback)
			}()
		}
	}()
}

func (l *LAN) update(api *uhppoted.UHPPOTED, id uint32, controller *Controller, callback Callback) {
	log.Printf("%v: refreshing LAN controller status", id)

	if info, err := api.GetDevice(uhppoted.GetDeviceRequest{DeviceID: uhppoted.DeviceID(id)}); err != nil {
		log.Printf("%v", err)
	} else if info == nil {
		log.Printf("Got %v response to get-device request for %v", info, id)
	} else {
		l.store(id, *info, controller)
	}

	if status, err := api.GetStatus(uhppoted.GetStatusRequest{DeviceID: uhppoted.DeviceID(id)}); err != nil {
		log.Printf("%v", err)
	} else if status == nil {
		log.Printf("Got %v response to get-status request for %v", status, id)
	} else {
		l.store(id, *status, controller)
	}

	if cards, err := api.GetCardRecords(uhppoted.GetCardRecordsRequest{DeviceID: uhppoted.DeviceID(id)}); err != nil {
		log.Printf("%v", err)
	} else if cards == nil {
		log.Printf("Got %v response to get-card-records request for %v", cards, id)
	} else {
		l.store(id, *cards, controller)
	}

	if events, err := api.GetEventRange(uhppoted.GetEventRangeRequest{DeviceID: uhppoted.DeviceID(id)}); err != nil {
		log.Printf("%v", err)
	} else if events == nil {
		log.Printf("Got %v response to get-event-range request for %v", events, id)
	} else {
		l.store(id, *events, controller)
	}

	for _, d := range []uint8{1, 2, 3, 4} {
		if delay, err := api.GetDoorDelay(uhppoted.GetDoorDelayRequest{DeviceID: uhppoted.DeviceID(id), Door: d}); err != nil {
			log.Printf("%v", err)
		} else if delay == nil {
			log.Printf("Got %v response to get-door-delay request for %v", delay, id)
		} else {
			l.store(id, *delay, controller)
		}
	}

	for _, d := range []uint8{1, 2, 3, 4} {
		if control, err := api.GetDoorControl(uhppoted.GetDoorControlRequest{DeviceID: uhppoted.DeviceID(id), Door: d}); err != nil {
			log.Printf("%v", err)
		} else if control == nil {
			log.Printf("Got %v response to get-door-control request for %v", control, id)
		} else {
			l.store(id, *control, controller)
		}
	}

	if recent, err := api.GetEvents(uhppoted.GetEventsRequest{DeviceID: uhppoted.DeviceID(id), Max: 5}); err != nil {
		log.Printf("%v", err)
	} else if callback != nil {
		callback.Append(id, recent.Events)
	}
}

func (l *LAN) store(id uint32, info interface{}, controller *Controller) {
	cache.guard.Lock()
	defer cache.guard.Unlock()

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
				oid := catalog.OID(door + ".2")
				catalog.PutV(oid, v.Delay, false)
			}
		}

	case uhppoted.GetDoorControlResponse:
		if controller != nil {
			if door, ok := controller.Doors[v.Door]; ok {
				oid := catalog.OID(door + ".3")
				catalog.PutV(oid, v.Control, false)
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
}

func (l *LAN) synchTime(controllers []*Controller) {
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
}

func (l *LAN) synchDoors(controllers []*Controller) {
	api := l.api(controllers)

	for _, c := range controllers {
		if c.DeviceID != nil {
			device := uhppoted.DeviceID(*c.DeviceID)

			// ... update door delays
			for _, door := range []uint8{1, 2, 3, 4} {
				if oid, ok := c.Doors[door]; ok && oid != "" {
					configured, modified := catalog.GetV(catalog.OID(oid).Append(catalog.DoorDelayConfigured))
					actual, _ := catalog.GetV(catalog.OID(oid).Append(catalog.DoorDelay))

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
							catalog.PutV(catalog.OID(oid).Append(catalog.DoorDelayConfigured), delay, false)
							catalog.PutV(catalog.OID(oid).Append(catalog.DoorDelay), delay, true)
							log.Printf("INFO  %v: synchronized door %v delay (%v)", response.DeviceID, door, delay)
						}
					}
				}
			}

			// ... update door control states
			for _, door := range []uint8{1, 2, 3, 4} {
				if oid, ok := c.Doors[door]; ok && oid != "" {
					configured, modified := catalog.GetV(catalog.OID(oid).Append(catalog.DoorControlConfigured))
					actual, _ := catalog.GetV(catalog.OID(oid).Append(catalog.DoorControl))

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
							catalog.PutV(catalog.OID(oid).Append(catalog.DoorControlConfigured), mode, false)
							catalog.PutV(catalog.OID(oid).Append(catalog.DoorControl), mode, true)
							log.Printf("INFO  %v: synchronized door %v control (%v)", response.DeviceID, door, mode)
						}
					}
				}
			}
		}
	}
}

func (l LAN) stash() {
}

func (l *LAN) log(auth auth.OpAuth, operation string, OID catalog.OID, field string, description string) {
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

	audit.Write(record)
}
