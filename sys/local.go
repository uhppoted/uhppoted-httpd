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
	BindAddress      *address `json:"bind-address"`
	BroadcastAddress *address `json:"broadcast-address"`
	ListenAddress    *address `json:"listen-address"`
	Debug            bool     `json:"debug"`
	devices          map[uint32]address
	api              uhppoted.UHPPOTED
	cache            map[uint32]device
	guard            sync.RWMutex
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
func (l *Local) Init(devices map[string]*Controller) {
	u := uhppote.UHPPOTE{
		BindAddress:      (*net.UDPAddr)(l.BindAddress),
		BroadcastAddress: (*net.UDPAddr)(l.BroadcastAddress),
		ListenAddress:    (*net.UDPAddr)(l.ListenAddress),
		Devices:          map[uint32]*uhppote.Device{},
		Debug:            l.Debug,
	}

	for _, v := range devices {
		l.devices[v.DeviceID] = v.IP

		addr := net.UDPAddr(v.IP)
		u.Devices[v.DeviceID] = &uhppote.Device{
			DeviceID: v.DeviceID,
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

func (l *Local) Controllers(controllers map[string]*Controller) []interface{} {
	type controller struct {
		ID         string
		created    time.Time
		Name       *types.Name
		DeviceID   uint32
		IP         ip
		SystemTime datetime
		Cards      *records
		Events     *records
		Doors      map[uint8]string
		Status     status
	}

	list := []controller{}

	for _, v := range controllers {
		c := controller{
			ID:       v.ID,
			created:  v.Created,
			Name:     v.Name,
			DeviceID: v.DeviceID,
			IP: ip{
				IP: &v.IP,
			},
			Doors: map[uint8]string{},
		}

		for _, d := range sys.Doors {
			if d.DeviceID == c.DeviceID {
				c.Doors[d.Door] = d.Name
			}
		}

		if cached, ok := l.cache[v.DeviceID]; ok {
			c.Cards = (*records)(cached.cards)
			c.Events = (*records)(cached.events)

			if cached.address != nil {
				if cached.address.Equal(c.IP.IP.IP) {
					c.IP.Status = StatusOk
				} else {
					c.IP.Status = StatusError
				}

				c.IP.IP = cached.address
			}

			if cached.datetime != nil {
				tz := time.Local
				if v.TimeZone != "" {
					if l, err := time.LoadLocation(v.TimeZone); err == nil {
						tz = l
					}
				}

				t := time.Time(*cached.datetime)
				T := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), tz)
				delta := math.Abs(time.Since(T).Round(time.Second).Seconds())

				if delta > WINDOW {
					c.SystemTime.Status = StatusError
				} else {
					c.SystemTime.Status = StatusOk
				}

				dt := types.DateTime(T)
				c.SystemTime.DateTime = &dt
				c.SystemTime.TimeZone = tz
			}

			switch dt := time.Now().Sub(cached.touched); {
			case dt < DeviceOk:
				c.Status = StatusOk
			case dt < DeviceUncertain:
				c.Status = StatusUncertain
			}
		}

		list = append(list, c)
	}

	sort.SliceStable(list, func(i, j int) bool { return list[i].created.Before(list[j].created) })

	// Ref. https://golang.org/doc/faq#convert_slice_of_interface
	result := make([]interface{}, len(list))
	for i, c := range list {
		result[i] = c
	}

	return result
}

// func (l *Local) ControllersX() []ControllerX {
// 	devices := map[uint32]ControllerX{}
// 	for k, v := range l.Devices {
// 		addr := v.IP // alias so that the loop below doesn't overwrite the configured value
//
// 		tz := time.Local
// 		if v.TimeZone != "" {
// 			if l, err := time.LoadLocation(v.TimeZone); err == nil {
// 				tz = l
// 			}
// 		}
//
// 		name := types.Name(v.Name)
//
// 		devices[k] = ControllerX{
// 			ID:       ID(k),
// 			created:  v.Created,
// 			Name:     &name,
// 			DeviceID: k,
// 			IP: ip{
// 				IP:     &addr,
// 				Status: StatusUnknown,
// 			},
// 			SystemTime: datetime{
// 				TimeZone: tz,
// 			},
// 			Doors:  map[uint8]string{},
// 			Status: StatusUnknown,
// 		}
// 	}
//
// 	list := []ControllerX{}
// 	for k, controller := range devices {
// 		if cached, ok := l.cache[k]; ok {
// 			controller.Cards = (*records)(cached.cards)
// 			controller.Events = (*records)(cached.events)
//
// 			if cached.address != nil {
// 				if cached.address.Equal(controller.IP.IP.IP) {
// 					controller.IP.Status = StatusOk
// 				} else {
// 					controller.IP.Status = StatusError
// 				}
//
// 				controller.IP.IP = cached.address
// 			}
//
// 			if cached.datetime != nil {
// 				tz := controller.SystemTime.TimeZone
// 				t := time.Time(*cached.datetime)
// 				T := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), tz)
// 				delta := math.Abs(time.Since(T).Round(time.Second).Seconds())
//
// 				if delta > WINDOW {
// 					controller.SystemTime.Status = StatusError
// 				} else {
// 					controller.SystemTime.Status = StatusOk
// 				}
//
// 				dt := types.DateTime(T)
// 				controller.SystemTime.DateTime = &dt
// 			}
//
// 			switch dt := time.Now().Sub(cached.touched); {
// 			case dt < DeviceOk:
// 				controller.Status = StatusOk
// 			case dt < DeviceUncertain:
// 				controller.Status = StatusUncertain
// 			}
// 		}
//
// 		list = append(list, controller)
// 	}
//
// 	// ... append the 'found but not configured' controllers
// 	for k, cached := range l.cache {
// 		if _, ok := devices[k]; !ok {
// 			controller := ControllerX{
// 				ID:       ID(k),
// 				created:  time.Now(),
// 				Name:     nil,
// 				DeviceID: k,
// 				IP: ip{
// 					IP:     cached.address,
// 					Status: StatusOk,
// 				},
// 				SystemTime: datetime{
// 					DateTime: cached.datetime,
// 					TimeZone: time.Local,
// 				},
// 				Cards:  (*records)(cached.cards),
// 				Events: (*records)(cached.events),
// 				Doors:  map[uint8]string{},
// 				Status: StatusUnknown,
// 			}
//
// 			if cached.datetime != nil {
// 				tz := controller.SystemTime.TimeZone
// 				t := time.Time(*cached.datetime)
// 				T := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), tz)
// 				delta := math.Abs(time.Since(T).Round(time.Second).Seconds())
//
// 				if delta > WINDOW {
// 					controller.SystemTime.Status = StatusError
// 				} else {
// 					controller.SystemTime.Status = StatusOk
// 				}
// 			}
//
// 			switch dt := time.Now().Sub(cached.touched); {
// 			case dt < DeviceOk:
// 				controller.Status = StatusUnconfigured
// 			case dt < DeviceUncertain:
// 				controller.Status = StatusUncertain
// 			default:
// 				controller.Status = StatusUnknown
// 			}
//
// 			list = append(list, controller)
// 		}
// 	}
//
// 	sort.SliceStable(list, func(i, j int) bool { return list[i].created.Before(list[j].created) })
//
// 	return list
// }

// func (l *Local) Controller(id uint32) *ControllerX {
// 	if v, ok := l.Devices[id]; ok {
// 		name := types.Name(v.Name)
// 		addr := v.IP // alias to avoid overwriting the configured value
//
// 		tz := time.Local
// 		if v.TimeZone != "" {
// 			if l, err := time.LoadLocation(v.TimeZone); err == nil {
// 				tz = l
// 			}
// 		}
//
// 		controller := ControllerX{
// 			ID:       ID(id),
// 			created:  v.Created,
// 			Name:     &name,
// 			DeviceID: id,
// 			IP: ip{
// 				IP:     &addr,
// 				Status: StatusUnknown,
// 			},
// 			SystemTime: datetime{
// 				TimeZone: tz,
// 			},
// 			Doors:  map[uint8]string{},
// 			Status: StatusUnknown,
// 		}
//
// 		if cached, ok := l.cache[id]; ok {
// 			controller.Cards = (*records)(cached.cards)
// 			controller.Events = (*records)(cached.events)
//
// 			if cached.address != nil {
// 				if cached.address.Equal(controller.IP.IP.IP) {
// 					controller.IP.Status = StatusOk
// 				} else {
// 					controller.IP.Status = StatusError
// 				}
//
// 				controller.IP.IP = cached.address
// 			}
//
// 			if cached.datetime != nil {
// 				tz := controller.SystemTime.TimeZone
// 				t := time.Time(*cached.datetime)
// 				T := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), tz)
// 				delta := math.Abs(time.Since(T).Round(time.Second).Seconds())
//
// 				if delta > WINDOW {
// 					controller.SystemTime.Status = StatusError
// 				} else {
// 					controller.SystemTime.Status = StatusOk
// 				}
//
// 				dt := types.DateTime(T)
// 				controller.SystemTime.DateTime = &dt
// 			}
//
// 			switch dt := time.Now().Sub(cached.touched); {
// 			case dt < DeviceOk:
// 				controller.Status = StatusOk
// 			case dt < DeviceUncertain:
// 				controller.Status = StatusUncertain
// 			}
// 		}
//
// 		return &controller
// 	}
//
// 	return nil
// }

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
	list := map[uint32]struct{}{}
	for k, _ := range l.devices {
		list[k] = struct{}{}
	}

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

				list[k] = struct{}{}
			}
		}

		for k, _ := range list {
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
		if d, ok := l.devices[id]; ok {
			port = d.Port
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

func ID(id uint32) string {
	return fmt.Sprintf("L%d", id)
}
