package interfaces

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	core "github.com/uhppoted/uhppote-core/types"
	"github.com/uhppoted/uhppote-core/uhppote"
	"github.com/uhppoted/uhppoted-httpd/audit"
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/db"
	"github.com/uhppoted/uhppoted-httpd/types"
	"github.com/uhppoted/uhppoted-lib/uhppoted"
)

type LANx struct {
	OID              catalog.OID
	Name             string
	BindAddress      core.BindAddr
	BroadcastAddress core.BroadcastAddr
	ListenAddress    core.ListenAddr
	Debug            bool

	created      types.DateTime
	deleted      *types.DateTime
	unconfigured bool
}

type Controller interface {
	OID() catalog.OID
	Name() string
	DeviceID() uint32
	EndPoint() *net.UDPAddr
	Door(uint8) (catalog.OID, bool)
}

var created = types.DateTimeNow()

func (l LANx) String() string {
	return fmt.Sprintf("%v", l.Name)
}

func (l *LANx) IsValid() bool {
	if l != nil {
		if strings.TrimSpace(l.Name) != "" {
			return true
		}
	}

	return false
}

func (l *LANx) IsDeleted() bool {
	if l != nil && l.deleted != nil {
		return true
	}

	return false
}

func (l *LANx) AsObjects() []interface{} {
	if l.deleted != nil {
		return []interface{}{
			catalog.NewObject2(l.OID, LANDeleted, l.deleted),
		}
	}

	objects := []interface{}{
		catalog.NewObject(l.OID, types.StatusUncertain),
		catalog.NewObject2(l.OID, LANStatus, l.status()),
		catalog.NewObject2(l.OID, LANCreated, l.created),
		catalog.NewObject2(l.OID, LANDeleted, l.deleted),
		catalog.NewObject2(l.OID, LANType, "LAN"),
		catalog.NewObject2(l.OID, LANName, l.Name),
		catalog.NewObject2(l.OID, LANBindAddress, l.BindAddress),
		catalog.NewObject2(l.OID, LANBroadcastAddress, l.BroadcastAddress),
		catalog.NewObject2(l.OID, LANListenAddress, l.ListenAddress),
	}

	return objects
}

func (l *LANx) AsRuleEntity() interface{} {
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

func (l *LANx) API(controllers []Controller) *uhppoted.UHPPOTED {
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
				Rollover: 100000,
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

func (l *LANx) set(auth auth.OpAuth, oid catalog.OID, value string, dbc db.DBC) ([]catalog.Object, error) {
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
	}

	return objects, nil
}

func (l *LANx) status() types.Status {
	return types.StatusOk
}

func (l LANx) Clone() LANx {
	return LANx{
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

// A long-running function i.e. expects to be invoked from an external goroutine
func (l *LANx) Refresh() {
}

func (l *LANx) Store(c Controller, info interface{}) {
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

		//	case acl.Diff:
		//		if len(v.Updated)+len(v.Added)+len(v.Deleted) > 0 {
		//			catalog.PutV(controller.OID(), ControllerCardsStatus, types.StatusError)
		//		} else {
		//			catalog.PutV(controller.OID(), ControllerCardsStatus, types.StatusOk)
		//		}
	}
}

func (l LANx) serialize() ([]byte, error) {
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

func (l *LANx) deserialize(bytes []byte) error {
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

// func (l *LANx) log(auth auth.OpAuth, operation string, OID catalog.OID, field string, description string, dbc db.DBC) {
// 	uid := ""
// 	if auth != nil {
// 		uid = auth.UID()
// 	}
//
// 	record := audit.AuditRecord{
// 		UID:       uid,
// 		OID:       OID,
// 		Component: "interface",
// 		Operation: operation,
// 		Details: audit.Details{
// 			ID:          "LAN",
// 			Name:        stringify(l.Name, ""),
// 			Field:       field,
// 			Description: description,
// 		},
// 	}
//
// 	if dbc != nil {
// 		dbc.Write(record)
// 	}
// }

func (l *LANx) log(auth auth.OpAuth, operation string, OID catalog.OID, field, description, before, after string, dbc db.DBC) {
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
