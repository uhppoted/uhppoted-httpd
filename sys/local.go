package system

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/uhppoted/uhppote-core/uhppote"
	"github.com/uhppoted/uhppoted-api/acl"
	"github.com/uhppoted/uhppoted-api/uhppoted"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type Local struct {
	BindAddress      *address              `json:"bind-address"`
	BroadcastAddress *address              `json:"broadcast-address"`
	ListenAddress    *address              `json:"listen-address"`
	Devices          map[uint32]controller `json:"controllers"`
	Debug            bool                  `json:"debug"`
	cache            map[uint32]device
	guard            sync.RWMutex
}

type controller struct {
	Created time.Time `json:"created"`
	Name    string    `json:"name"`
	IP      address   `json:"IPv4"`
}

type device struct {
	touched  time.Time
	address  *address
	datetime *types.DateTime
	cards    *uint32
	events   *uint32
}

const (
	DeviceOk        = 10 * time.Second
	DeviceUncertain = 20 * time.Second
)

func (l *Local) Controllers() []Controller {
	list := []Controller{}

	for k, v := range l.Devices {
		controller := Controller{
			created: v.Created,
			Name:    v.Name,
			ID:      k,
			Doors:   map[uint8]string{},
			Status:  StatusUnknown,
		}

		if cached, ok := l.cache[k]; ok {
			controller.IP = cached.address
			controller.DateTime = cached.datetime
			controller.Cards = cached.cards
			controller.Events = cached.events

			switch dt := time.Now().Sub(cached.touched); {
			case dt < DeviceOk:
				controller.Status = StatusOk
			case dt < DeviceUncertain:
				controller.Status = StatusUncertain
			}
		}

		list = append(list, controller)
	}

	sort.SliceStable(list, func(i, j int) bool { return list[i].created.Before(list[j].created) })

	return list
}

func (l *Local) Update(permissions []types.Permissions) {
	log.Printf("Updating ACL")

	access, err := consolidate(permissions)
	if err != nil {
		warn(err)
		return
	}

	if access == nil {
		warn(fmt.Errorf("Invalid ACL from permissions: %v", access))
		return
	}

	// TODO: move to local.Init()
	u := uhppote.UHPPOTE{
		BindAddress:      (*net.UDPAddr)(l.BindAddress),
		BroadcastAddress: (*net.UDPAddr)(l.BroadcastAddress),
		ListenAddress:    (*net.UDPAddr)(l.ListenAddress),
		Devices:          map[uint32]*uhppote.Device{},
		Debug:            l.Debug,
	}

	for k, v := range l.Devices {
		addr := net.UDPAddr(v.IP)
		u.Devices[k] = &uhppote.Device{
			DeviceID: k,
			Address:  &addr,
			Rollover: 100000,
			Doors:    []string{},
		}
	}
	// TODO END

	rpt, err := acl.PutACL(&u, *access, false)
	if err != nil {
		warn(err)
		return
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

func (l *Local) refresh() {
	for k, _ := range l.Devices {
		id := k
		go func() {
			l.update(id)
		}()
	}
}

func (l *Local) update(id uint32) {
	log.Printf("%v: refreshing 'local' controller status", id)

	// TODO: move to local.Init()
	u := uhppote.UHPPOTE{
		BindAddress:      (*net.UDPAddr)(l.BindAddress),
		BroadcastAddress: (*net.UDPAddr)(l.BroadcastAddress),
		ListenAddress:    (*net.UDPAddr)(l.ListenAddress),
		Devices:          map[uint32]*uhppote.Device{},
		Debug:            l.Debug,
	}

	for k, v := range l.Devices {
		addr := net.UDPAddr(v.IP)
		u.Devices[k] = &uhppote.Device{
			DeviceID: k,
			Address:  &addr,
			Rollover: 100000,
			Doors:    []string{},
		}
	}

	logger := log.New(os.Stdout, "local", log.LstdFlags|log.LUTC)
	impl := uhppoted.UHPPOTED{
		Uhppote: &u,
		Log:     logger,
	}

	// TODO END

	if info, err := impl.GetDevice(uhppoted.GetDeviceRequest{DeviceID: uhppoted.DeviceID(id)}); err != nil {
		log.Printf("%v", err)
	} else if info == nil {
		log.Printf("Got %v response to get-device request for %v", info, id)
	} else {
		l.store(id, *info)
	}

	if status, err := impl.GetStatus(uhppoted.GetStatusRequest{DeviceID: uhppoted.DeviceID(id)}); err != nil {
		log.Printf("%v", err)
	} else if status == nil {
		log.Printf("Got %v response to get-status request for %v", status, id)
	} else {
		l.store(id, *status)
	}

	if cards, err := impl.GetCardRecords(uhppoted.GetCardRecordsRequest{DeviceID: uhppoted.DeviceID(id)}); err != nil {
		log.Printf("%v", err)
	} else if cards == nil {
		log.Printf("Got %v response to get-card-records request for %v", cards, id)
	} else {
		l.store(id, *cards)
	}

	if events, err := impl.GetEventRange(uhppoted.GetEventRangeRequest{DeviceID: uhppoted.DeviceID(id)}); err != nil {
		log.Printf("%v", err)
	} else if events == nil {
		log.Printf("Got %v response to get-event-range request for %v", events, id)
	} else {
		l.store(id, *events)
	}
}

func (l *Local) store(id uint32, info interface{}) {
	l.guard.Lock()

	defer l.guard.Unlock()

	if l.cache == nil {
		l.cache = map[uint32]device{}
	}

	cached, ok := l.cache[id]
	if !ok {
		cached = device{}
	}

	cached.touched = time.Now()

	switch v := info.(type) {
	case uhppoted.GetDeviceResponse:
		port := 60000
		if d, ok := l.Devices[id]; ok {
			port = d.IP.Port
		}

		addr := address(net.UDPAddr{
			IP:   v.IpAddress,
			Port: port,
		})
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
	}

	l.cache[id] = cached
}
