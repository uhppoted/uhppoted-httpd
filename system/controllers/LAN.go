package controllers

import (
	"fmt"
	"log"
	"sync"
	"time"

	core "github.com/uhppoted/uhppote-core/types"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/interfaces"
	"github.com/uhppoted/uhppoted-httpd/types"
	"github.com/uhppoted/uhppoted-lib/uhppoted"
)

type LAN struct {
	interfaces.LANx
}

func (l *LAN) String() string {
	return fmt.Sprintf("%v", l.Name)
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

// Long-running function. Synchronous - expects to be invoked from an external goroutine.
func (l *LAN) refresh(controllers []*Controller, callback Callback) {
	l.Refresh()

	api := l.api(controllers)

	var wg sync.WaitGroup

	for _, c := range controllers {
		if c.deviceID != nil && *c.deviceID != 0 && c.deleted == nil {
			wg.Add(1)

			controller := c

			go func(v *Controller) {
				defer wg.Done()

				l.Update(api, v)

				deviceID := v.DeviceID()
				recent, err := api.GetEvents(uhppoted.GetEventsRequest{DeviceID: uhppoted.DeviceID(deviceID), Max: 5})
				if err != nil {
					log.Printf("%v", err)
				} else if callback != nil {
					callback.Append(v.DeviceID(), recent.Events)
				}
			}(controller)
		}
	}

	wg.Wait()
}

func (l *LAN) synchTime(controllers []*Controller) []catalog.Object {
	objects := []catalog.Object{}
	api := l.api(controllers)
	for _, c := range controllers {
		if c.deviceID != nil {
			device := uhppoted.DeviceID(*c.deviceID)
			location := time.Local

			if c.TimeZone != nil {
				timezone := *c.TimeZone
				if tz, err := types.Timezone(timezone); err == nil && tz != nil {
					location = tz
				}
			}

			now := time.Now().In(location)
			datetime := core.DateTime(now)

			request := uhppoted.SetTimeRequest{
				DeviceID: device,
				DateTime: datetime,
			}

			if response, err := api.SetTime(request); err != nil {
				log.Printf("ERROR %v", err)
			} else if response != nil {
				log.Printf("INFO  synchronized device-time %v %v", response.DeviceID, response.DateTime)
			}
		}
	}

	return objects
}

func (l *LAN) synchDoors(controllers []*Controller) []catalog.Object {
	objects := []catalog.Object{}
	api := l.api(controllers)

	for _, c := range controllers {
		if c.deviceID != nil {
			device := uhppoted.DeviceID(*c.deviceID)

			// ... update door delays
			for _, door := range []uint8{1, 2, 3, 4} {
				if oid, ok := c.Doors[door]; ok && oid != "" {
					configured := catalog.GetV(oid, catalog.DoorDelayConfigured)
					actual := catalog.GetV(oid, catalog.DoorDelay)
					modified := false

					if v := catalog.GetV(oid, catalog.DoorDelayModified); v != nil {
						if b, ok := v.(bool); ok {
							modified = b
						}
					}

					if configured != nil && (actual == nil || actual != configured) && modified {
						delay := configured.(uint8)

						request := uhppoted.SetDoorDelayRequest{
							DeviceID: device,
							Door:     door,
							Delay:    delay,
						}

						if response, err := api.SetDoorDelay(request); err != nil {
							log.Printf("ERROR %v", err)
						} else if response != nil {
							objects = append(objects, catalog.NewObject2(oid, DoorDelay, delay))
							objects = append(objects, catalog.NewObject2(oid, DoorDelayModified, false))
							log.Printf("INFO  %v: synchronized door %v delay (%v)", response.DeviceID, door, delay)
						}
					}
				}
			}

			// ... update door control states
			for _, door := range []uint8{1, 2, 3, 4} {
				if oid, ok := c.Doors[door]; ok && oid != "" {
					configured := catalog.GetV(oid, catalog.DoorControlConfigured)
					actual := catalog.GetV(oid, catalog.DoorControl)
					modified := false

					if v := catalog.GetV(oid, catalog.DoorControlModified); v != nil {
						if b, ok := v.(bool); ok {
							modified = b
						}
					}

					if configured != nil && (actual == nil || actual != configured) && modified {
						mode := configured.(core.ControlState)

						request := uhppoted.SetDoorControlRequest{
							DeviceID: device,
							Door:     door,
							Control:  mode,
						}

						if response, err := api.SetDoorControl(request); err != nil {
							log.Printf("ERROR %v", err)
						} else if response != nil {
							objects = append(objects, catalog.NewObject2(oid, DoorControl, mode))
							objects = append(objects, catalog.NewObject2(oid, DoorControlModified, false))
							log.Printf("INFO  %v: synchronized door %v control (%v)", response.DeviceID, door, mode)
						}
					}
				}
			}
		}
	}

	return objects
}
