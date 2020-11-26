package system

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sort"
	"time"

	"github.com/uhppoted/uhppote-core/uhppote"
	uhppoted "github.com/uhppoted/uhppoted-api/acl"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type Local struct {
	BindAddress      *address              `json:"bind-address"`
	BroadcastAddress *address              `json:"broadcast-address"`
	ListenAddress    *address              `json:"listen-address"`
	Devices          map[uint32]controller `json:"controllers"`
	Debug            bool                  `json:"debug"`
}

type controller struct {
	Created time.Time `json:"created"`
	Name    string    `json:"name"`
	IP      address   `json:"IPv4"`
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
		list = append(list, Controller{
			created: v.Created,
			Name:    v.Name,
			ID:      k,
			IP:      v.IP.String(),
		})
	}

	sort.SliceStable(list, func(i, j int) bool { return list[i].created.Before(list[j].created) })

	return list
}

func (l *Local) Update(permissions []types.Permissions) {
	log.Printf("Updating ACL")

	acl, err := consolidate(permissions)
	if err != nil {
		warn(err)
		return
	}

	if acl == nil {
		warn(fmt.Errorf("Invalid ACL from permissions: %v", acl))
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

	rpt, err := uhppoted.PutACL(&u, *acl, false)
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
