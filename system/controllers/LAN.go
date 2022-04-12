package controllers

import (
	"sync"

	"github.com/uhppoted/uhppoted-httpd/system/interfaces"
	"github.com/uhppoted/uhppoted-lib/acl"
)

type LAN struct {
	interfaces interfaces.Interfaces
	lan        interfaces.LAN
}

func (l *LAN) search(controllers []*Controller) ([]uint32, error) {
	devices := []interfaces.Controller{}
	for _, c := range controllers {
		if c.realized() {
			devices = append(devices, c)
		}
	}

	return l.lan.Search(devices)
}

func (l *LAN) refresh(controllers []*Controller) {
	realized := []interfaces.Controller{}

	for _, c := range controllers {
		if c.realized() {
			realized = append(realized, c)
		}
	}

	l.interfaces.Refresh(realized)
	l.interfaces.GetEvents(realized)
}

func (l *LAN) synchTime(controllers []*Controller) {
	f := func(l *LAN, c *Controller) {
		l.lan.SynchTime(c)
	}

	l.exec(controllers, f)
}

func (l *LAN) synchDoors(controllers []*Controller) {
	f := func(l *LAN, c *Controller) {
		l.lan.SynchDoors(c)
	}

	l.exec(controllers, f)
}

func (l *LAN) exec(controllers []*Controller, f func(l *LAN, c *Controller)) {
	var wg sync.WaitGroup

	for _, c := range controllers {
		if c.realized() {
			wg.Add(1)

			controller := c

			go func(v *Controller) {
				defer wg.Done()

				f(l, v)
			}(controller)
		}
	}

	wg.Wait()
}

func (l *LAN) compare(controllers []*Controller, permissions acl.ACL) error {
	devices := []interfaces.Controller{}
	for _, c := range controllers {
		if c.realized() {
			devices = append(devices, c)
		}
	}

	return l.lan.CompareACL(devices, permissions)
}

func (l *LAN) update(controllers []*Controller, permissions acl.ACL) error {
	devices := []interfaces.Controller{}
	for _, c := range controllers {
		if c.realized() {
			devices = append(devices, c)
		}
	}

	return l.lan.UpdateACL(devices, permissions)
}

// type icontroller struct {
// 	oid      schema.OID
// 	name     string
// 	id       uint32
// 	endpoint *net.UDPAddr
// 	timezone *time.Location
// 	doors    map[uint8]schema.OID
// }

// func makeIController(c Controller) icontroller {
// 	doors := map[uint8]schema.OID{}

// 	for _, d := range []uint8{1, 2, 3, 4} {
// 		if oid, ok := c.Doors[d]; ok {
// 			doors[d] = oid
// 		}
// 	}

// 	return icontroller{
// 		oid:      c.OID,
// 		name:     c.name,
// 		id:       c.DeviceID,
// 		endpoint: c.EndPoint(),
// 		timezone: c.TimeZone(),
// 		doors:    doors,
// 	}
// }

// func (c icontroller) OIDx() schema.OID {
// 	return c.oid
// }

// func (c icontroller) Name() string {
// 	return c.name
// }

// func (c icontroller) ID() uint32 {
// 	return c.id
// }

// func (c icontroller) EndPoint() *net.UDPAddr {
// 	return c.endpoint
// }

// func (c icontroller) TimeZone() *time.Location {
// 	return c.timezone
// }

// func (c icontroller) Door(d uint8) (schema.OID, bool) {
// 	oid, ok := c.doors[d]

// 	return oid, ok
// }
