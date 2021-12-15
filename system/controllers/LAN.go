package controllers

import (
	"sync"

	"github.com/uhppoted/uhppoted-httpd/system/interfaces"
	"github.com/uhppoted/uhppoted-lib/acl"
)

type LAN struct {
	interfaces.LAN
}

func (l LAN) clone() LAN {
	return LAN{
		l.LAN,
	}
}

func (l *LAN) search(controllers []*Controller) ([]uint32, error) {
	devices := []interfaces.Controller{}
	for _, c := range controllers {
		if c.deviceID != 0 && c.deleted == nil {
			devices = append(devices, c)
		}
	}

	return l.Search(devices)
}

func (l *LAN) refresh(controllers []*Controller) {
	f := func(l *LAN, c *Controller) {
		l.Refresh(c)
	}

	l.exec(controllers, f)
}

func (l *LAN) synchTime(controllers []*Controller) {
	f := func(l *LAN, c *Controller) {
		l.SynchTime(c)
	}

	l.exec(controllers, f)
}

func (l *LAN) synchDoors(controllers []*Controller) {
	f := func(l *LAN, c *Controller) {
		l.SynchDoors(c)
	}

	l.exec(controllers, f)
}

func (l *LAN) exec(controllers []*Controller, f func(l *LAN, c *Controller)) {
	var wg sync.WaitGroup

	for _, c := range controllers {
		if c.deviceID != 0 && c.deleted == nil {
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
		if c.deviceID != 0 && c.deleted == nil {
			devices = append(devices, c)
		}
	}

	return l.CompareACL(devices, permissions)
}

func (l *LAN) update(controllers []*Controller, permissions acl.ACL) error {
	devices := []interfaces.Controller{}
	for _, c := range controllers {
		if c.deviceID != 0 && c.deleted == nil {
			devices = append(devices, c)
		}
	}

	return l.UpdateACL(devices, permissions)
}
