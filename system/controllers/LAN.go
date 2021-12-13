package controllers

import (
	"sync"

	"github.com/uhppoted/uhppoted-httpd/system/interfaces"
	"github.com/uhppoted/uhppoted-lib/uhppoted"
)

type LAN struct {
	interfaces.LANx
}

func (l LAN) clone() LAN {
	return LAN{
		l.LANx,
	}
}

func (l *LAN) api(controllers []*Controller) *uhppoted.UHPPOTED {
	devices := []interfaces.Controller{}
	for _, c := range controllers {
		if c.deviceID != nil && *c.deviceID != 0 && c.deleted == nil {
			devices = append(devices, c)
		}
	}

	return l.API(devices)
}

func (l *LAN) search(controllers []*Controller) ([]uint32, error) {
	devices := []interfaces.Controller{}
	for _, c := range controllers {
		if c.deviceID != nil && *c.deviceID != 0 && c.deleted == nil {
			devices = append(devices, c)
		}
	}

	return l.Search(devices)
}

// Synchronous long-running function - expects to be invoked from an external goroutine.
func (l *LAN) refresh(controllers []*Controller) {
	f := func(l *LAN, api *uhppoted.UHPPOTED, c *Controller) {
		l.Refresh(api, c)
	}

	l.exec(controllers, f)
}

// Synchronous long-running function - expects to be invoked from an external goroutine.
func (l *LAN) synchTime(controllers []*Controller) {
	f := func(l *LAN, api *uhppoted.UHPPOTED, c *Controller) {
		l.SynchTime(api, c)
	}

	l.exec(controllers, f)
}

// Synchronous long-running function - expects to be invoked from an external goroutine.
func (l *LAN) synchDoors(controllers []*Controller) {
	f := func(l *LAN, api *uhppoted.UHPPOTED, c *Controller) {
		l.SynchDoors(api, c)
	}

	l.exec(controllers, f)
}

func (l *LAN) exec(controllers []*Controller, f func(l *LAN, api *uhppoted.UHPPOTED, c *Controller)) {
	api := l.api(controllers)

	var wg sync.WaitGroup

	for _, c := range controllers {
		if c.deviceID != nil && *c.deviceID != 0 && c.deleted == nil {
			wg.Add(1)

			controller := c

			go func(v *Controller) {
				defer wg.Done()

				f(l, api, v)
			}(controller)
		}
	}

	wg.Wait()
}
