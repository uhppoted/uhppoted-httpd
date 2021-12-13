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
	api := l.api(controllers)

	var wg sync.WaitGroup

	for _, c := range controllers {
		if c.deviceID != nil && *c.deviceID != 0 && c.deleted == nil {
			wg.Add(1)

			controller := c

			go func(v *Controller) {
				defer wg.Done()

				l.Refresh(api, v)
			}(controller)
		}
	}

	wg.Wait()
}

// Synchronous long-running function - expects to be invoked from an external goroutine.
func (l *LAN) synchTime(controllers []*Controller) {
	api := l.api(controllers)

	list := []interfaces.Controller{}
	for _, c := range controllers {
		if c.deviceID != nil && *c.deviceID != 0 && c.deleted == nil {
			list = append(list, c)
		}
	}

	l.SynchTime(api, list)
}

// Synchronous long-running function - expects to be invoked from an external goroutine.
func (l *LAN) synchDoors(controllers []*Controller) {
	api := l.api(controllers)

	list := []interfaces.Controller{}
	for _, c := range controllers {
		if c.deviceID != nil && *c.deviceID != 0 && c.deleted == nil {
			list = append(list, c)
		}
	}

	l.SynchDoors(api, list)
}
