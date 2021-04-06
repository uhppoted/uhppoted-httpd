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
	"github.com/uhppoted/uhppoted-api/acl"
	"github.com/uhppoted/uhppoted-api/uhppoted"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type LAN struct {
	OID              string         `json:"OID"`
	Name             string         `json:"name"`
	BindAddress      *types.Address `json:"bind-address"`
	BroadcastAddress *types.Address `json:"broadcast-address"`
	ListenAddress    *types.Address `json:"listen-address"`
	Debug            bool           `json:"debug"`
}

type deviceCache struct {
	cache map[uint32]device
	guard sync.RWMutex
}

type device struct {
	touched  time.Time
	address  *types.Address
	datetime *types.DateTime
	cards    *uint32
	events   *uint32
	acl      status
}

const (
	DeviceOk        = 10 * time.Second
	DeviceUncertain = 20 * time.Second
)

const WINDOW = 300 // 5 minutes

var cache = deviceCache{
	cache: map[uint32]device{},
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
		}

		return &lan
	}

	return nil
}

func (l *LAN) api(controllers []*Controller) *uhppoted.UHPPOTED {
	u := uhppote.UHPPOTE{
		BindAddress:      (*net.UDPAddr)(l.BindAddress),
		BroadcastAddress: (*net.UDPAddr)(l.BroadcastAddress),
		ListenAddress:    (*net.UDPAddr)(l.ListenAddress),
		Devices:          map[uint32]*uhppote.Device{},
		Debug:            l.Debug,
	}

	for _, v := range controllers {
		if v.DeviceID == nil || *v.DeviceID == 0 || v.IP == nil {
			continue
		}

		name := v.Name.String()
		id := *v.DeviceID
		addr := net.UDPAddr(*v.IP)

		u.Devices[id] = &uhppote.Device{
			Name:     name,
			DeviceID: id,
			Address:  &addr,
			Rollover: 100000,
			Doors:    []string{},
			TimeZone: time.Local,
		}
	}

	api := uhppoted.UHPPOTED{
		Uhppote: &u,
		Log:     log.New(os.Stdout, "", log.LstdFlags|log.LUTC),
	}

	return &api
}

func (l *LAN) updateACL(controllers []*Controller, permissions acl.ACL) {
	log.Printf("Updating ACL")

	api := l.api(controllers)
	rpt, errors := acl.PutACL(api.Uhppote, permissions, false)
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

	devices := []*uhppote.Device{}
	api := l.api(controllers)
	for _, v := range api.Uhppote.DeviceList() {
		devices = append(devices, v)
	}

	current, errors := acl.GetACL(api.Uhppote, devices)
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
		l.store(id, compare[id])
	}

	return nil
}

func (l *LAN) refresh(controllers []*Controller) {
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
				if d, ok := api.Uhppote.DeviceList()[k]; ok {
					d.Address.IP = v.Address
					d.Address.Port = v.Port
				}

				list[k] = struct{}{}
			}
		}

		for k, _ := range list {
			id := k
			go func() {
				l.update(api, id)
			}()
		}
	}()
}

func (l *LAN) update(api *uhppoted.UHPPOTED, id uint32) {
	log.Printf("%v: refreshing LAN controller status", id)

	if info, err := api.GetDevice(uhppoted.GetDeviceRequest{DeviceID: uhppoted.DeviceID(id)}); err != nil {
		log.Printf("%v", err)
	} else if info == nil {
		log.Printf("Got %v response to get-device request for %v", info, id)
	} else {
		l.store(id, *info)
	}

	if status, err := api.GetStatus(uhppoted.GetStatusRequest{DeviceID: uhppoted.DeviceID(id)}); err != nil {
		log.Printf("%v", err)
	} else if status == nil {
		log.Printf("Got %v response to get-status request for %v", status, id)
	} else {
		l.store(id, *status)
	}

	if cards, err := api.GetCardRecords(uhppoted.GetCardRecordsRequest{DeviceID: uhppoted.DeviceID(id)}); err != nil {
		log.Printf("%v", err)
	} else if cards == nil {
		log.Printf("Got %v response to get-card-records request for %v", cards, id)
	} else {
		l.store(id, *cards)
	}

	if events, err := api.GetEventRange(uhppoted.GetEventRangeRequest{DeviceID: uhppoted.DeviceID(id)}); err != nil {
		log.Printf("%v", err)
	} else if events == nil {
		log.Printf("Got %v response to get-event-range request for %v", events, id)
	} else {
		l.store(id, *events)
	}
}

func (l *LAN) delete(c Controller) {
	cache.guard.Lock()
	defer cache.guard.Unlock()

	if l != nil && c.DeviceID != nil && *c.DeviceID != 0 {
		delete(cache.cache, *c.DeviceID)
	}
}

func (l *LAN) store(id uint32, info interface{}) {
	cache.guard.Lock()
	defer cache.guard.Unlock()

	cached, ok := cache.cache[id]
	if !ok {
		cached = device{}
	}

	cached.touched = time.Now()

	switch v := info.(type) {
	case uhppoted.GetDeviceResponse:
		addr := types.Address(v.Address)
		cached.address = &addr

	case uhppoted.GetStatusResponse:
		datetime := types.DateTime(v.Status.SystemDateTime)
		cached.datetime = &datetime

	case uhppoted.GetCardRecordsResponse:
		cards := v.Cards
		cached.cards = &cards

	case uhppoted.GetEventRangeResponse:
		events := v.Events.Last
		cached.events = events

	case acl.Diff:
		if len(v.Updated)+len(v.Added)+len(v.Deleted) > 0 {
			cached.acl = StatusError
		} else {
			cached.acl = StatusOk
		}
	}

	cache.cache[id] = cached
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
				log.Printf("INFO  sychronized device-time %v %v", response.DeviceID, response.DateTime)
			}
		}
	}
}
