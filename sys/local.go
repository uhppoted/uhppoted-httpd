package system

import (
	"bytes"
	"fmt"
	"log"
	"math"
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
	api              uhppoted.UHPPOTED
	cache            map[uint32]device
	guard            sync.RWMutex
}

type controller struct {
	Created  time.Time `json:"created"`
	Name     string    `json:"name"`
	IP       address   `json:"IPv4"`
	TimeZone string    `json:"timezone"`
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

const WINDOW = 300 // 5 minutes

// TODO (?) Move into custom JSON Unmarshal
//          Ref. http://choly.ca/post/go-json-marshalling/
func (l *Local) Init() {
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

	l.api = uhppoted.UHPPOTED{
		Uhppote: &u,
		Log:     log.New(os.Stdout, "local", log.LstdFlags|log.LUTC),
	}
}

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
			controller.Cards = cached.cards
			controller.Events = cached.events

			if cached.address == nil {
				controller.IP.IP = &v.IP
				controller.IP.Status = StatusUnknown
			} else {
				controller.IP.IP = cached.address
				if cached.address.Equal(v.IP.IP) {
					controller.IP.Status = StatusOk
				} else {
					controller.IP.Status = StatusError
				}
			}

			if cached.datetime == nil {
				controller.SystemTime.DateTime = nil
				controller.SystemTime.Status = StatusUnknown
			} else {
				location := time.Local
				if v.TimeZone != "" {
					if l, err := time.LoadLocation(v.TimeZone); err == nil {
						location = l
					}
				}

				t := time.Time(*cached.datetime)
				T := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), location)
				delta := math.Abs(time.Since(T).Round(time.Second).Seconds())

				if delta > WINDOW {
					controller.SystemTime.Status = StatusError
				} else {
					controller.SystemTime.Status = StatusOk
				}

				dt := types.DateTime(T)
				controller.SystemTime.DateTime = &dt
			}

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

	rpt, err := acl.PutACL(l.api.Uhppote, *access, false)
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
	go func() {
		if devices, err := l.api.GetDevices(uhppoted.GetDevicesRequest{}); err != nil {
			log.Printf("%v", err)
		} else if devices == nil {
			log.Printf("Got %v response to get-devices request", devices)
		} else {
			for k, v := range devices.Devices {
				if d, ok := l.api.Uhppote.Devices[k]; ok {
					d.Address.IP = v.Address
					d.Address.Port = 60000
				}
			}
		}

		for k, _ := range l.Devices {
			id := k
			go func() {
				l.update(id)
			}()
		}
	}()
}

func (l *Local) update(id uint32) {
	log.Printf("%v: refreshing 'local' controller status", id)

	if info, err := l.api.GetDevice(uhppoted.GetDeviceRequest{DeviceID: uhppoted.DeviceID(id)}); err != nil {
		log.Printf("%v", err)
	} else if info == nil {
		log.Printf("Got %v response to get-device request for %v", info, id)
	} else {
		l.store(id, *info)
	}

	if status, err := l.api.GetStatus(uhppoted.GetStatusRequest{DeviceID: uhppoted.DeviceID(id)}); err != nil {
		log.Printf("%v", err)
	} else if status == nil {
		log.Printf("Got %v response to get-status request for %v", status, id)
	} else {
		l.store(id, *status)
	}

	if cards, err := l.api.GetCardRecords(uhppoted.GetCardRecordsRequest{DeviceID: uhppoted.DeviceID(id)}); err != nil {
		log.Printf("%v", err)
	} else if cards == nil {
		log.Printf("Got %v response to get-card-records request for %v", cards, id)
	} else {
		l.store(id, *cards)
	}

	if events, err := l.api.GetEventRange(uhppoted.GetEventRangeRequest{DeviceID: uhppoted.DeviceID(id)}); err != nil {
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
