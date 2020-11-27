package system

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
	datetime *types.DateTime
}

type address net.UDPAddr

func (a *address) String() string {
	return (*net.UDPAddr)(a).String()
}

func (a *address) UnmarshalJSON(bytes []byte) error {
	var s string

	if err := json.Unmarshal(bytes, &s); err != nil {
		return err
	}

	addr, err := net.ResolveUDPAddr("udp", s)
	if err != nil {
		return err
	}

	*a = address(*addr)

	return nil
}

func (l *Local) Controllers() []Controller {
	list := []Controller{}

	for k, v := range l.Devices {
		controller := Controller{
			created: v.Created,
			Name:    v.Name,
			ID:      k,
			IP:      v.IP.String(),
		}

		if state, ok := l.cache[k]; ok {
			controller.DateTime = state.datetime
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
		l.update(k)
	}
}

func (l *Local) update(id uint32) {
	go func() {
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

		rq := uhppoted.GetStatusRequest{
			DeviceID: uhppoted.DeviceID(id),
		}

		status, err := impl.GetStatus(rq)
		if err != nil {
			log.Printf("%v", err)
		} else if status == nil {
			log.Printf("Got %v response to get-status request for %v", status, id)
		} else {
			l.store(id, *status)
		}
	}()
}

func (l *Local) store(id uint32, status uhppoted.GetStatusResponse) {
	datetime := types.DateTime(status.Status.SystemDateTime)

	l.guard.Lock()

	defer l.guard.Unlock()

	if l.cache == nil {
		l.cache = map[uint32]device{}
	}

	cached, ok := l.cache[id]
	if !ok {
		cached = device{}
	}

	cached.datetime = &datetime

	l.cache[id] = cached
}
