package controllers

import (
	"fmt"
	"log"
	"sync"
	"time"

	core "github.com/uhppoted/uhppote-core/types"
	"github.com/uhppoted/uhppoted-httpd/audit"
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/db"
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
	for _, v := range controllers {
		devices = append(devices, v)
	}

	return l.API(devices)
}

// Possibly a long-running function - expects to be invoked from an external goroutine
func (l *LAN) refresh(controllers []*Controller, callback Callback) {
	l.Refresh()

	list := map[uint32]struct{}{}
	for _, c := range controllers {
		if c.deviceID != nil && *c.deviceID != 0 {
			list[*c.deviceID] = struct{}{}
		}
	}

	api := l.api(controllers)
	if devices, err := api.GetDevices(uhppoted.GetDevicesRequest{}); err != nil {
		log.Printf("%v", err)
	} else if devices == nil {
		log.Printf("Got %v response to get-devices request", devices)
	} else {
		for k, v := range devices.Devices {
			if d, ok := api.UHPPOTE.DeviceList()[k]; ok {
				d.Address.IP = v.Address
				d.Address.Port = v.Port
			}

			list[k] = struct{}{}
		}
	}

	var wg sync.WaitGroup

	for k, _ := range list {
		wg.Add(1)

		id := k
		go func() {
			defer wg.Done()

			var controller *Controller
			for _, c := range controllers {
				if c.deviceID != nil && *c.deviceID == id {
					controller = c
					break
				}
			}

			l.update(api, id, controller, callback)
		}()
	}

	wg.Wait()
}

func (l *LAN) update(api *uhppoted.UHPPOTED, id uint32, controller *Controller, callback Callback) {
	log.Printf("%v: refreshing LAN controller status", id)

	if info, err := api.GetDevice(uhppoted.GetDeviceRequest{DeviceID: uhppoted.DeviceID(id)}); err != nil {
		log.Printf("%v", err)
	} else if info == nil {
		log.Printf("Got %v response to get-device request for %v", info, id)
	} else {
		l.Store(controller, *info)
	}

	if status, err := api.GetStatus(uhppoted.GetStatusRequest{DeviceID: uhppoted.DeviceID(id)}); err != nil {
		log.Printf("%v", err)
	} else if status == nil {
		log.Printf("Got %v response to get-status request for %v", status, id)
	} else {
		l.Store(controller, *status)
	}

	if cards, err := api.GetCardRecords(uhppoted.GetCardRecordsRequest{DeviceID: uhppoted.DeviceID(id)}); err != nil {
		log.Printf("%v", err)
	} else if cards == nil {
		log.Printf("Got %v response to get-card-records request for %v", cards, id)
	} else {
		l.Store(controller, *cards)
	}

	if events, err := api.GetEventRange(uhppoted.GetEventRangeRequest{DeviceID: uhppoted.DeviceID(id)}); err != nil {
		log.Printf("%v", err)
	} else if events == nil {
		log.Printf("Got %v response to get-event-range request for %v", events, id)
	} else {
		l.Store(controller, *events)
	}

	for _, d := range []uint8{1, 2, 3, 4} {
		if delay, err := api.GetDoorDelay(uhppoted.GetDoorDelayRequest{DeviceID: uhppoted.DeviceID(id), Door: d}); err != nil {
			log.Printf("%v", err)
		} else if delay == nil {
			log.Printf("Got %v response to get-door-delay request for %v", delay, id)
		} else {
			l.Store(controller, *delay)
		}
	}

	for _, d := range []uint8{1, 2, 3, 4} {
		if control, err := api.GetDoorControl(uhppoted.GetDoorControlRequest{DeviceID: uhppoted.DeviceID(id), Door: d}); err != nil {
			log.Printf("%v", err)
		} else if control == nil {
			log.Printf("Got %v response to get-door-control request for %v", control, id)
		} else {
			l.Store(controller, *control)
		}
	}

	if recent, err := api.GetEvents(uhppoted.GetEventsRequest{DeviceID: uhppoted.DeviceID(id), Max: 5}); err != nil {
		log.Printf("%v", err)
	} else if callback != nil {
		callback.Append(id, recent.Events)
	}
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

func (l *LAN) log(auth auth.OpAuth, operation string, OID catalog.OID, field string, description string, dbc db.DBC) {
	uid := ""
	if auth != nil {
		uid = auth.UID()
	}

	record := audit.AuditRecord{
		UID:       uid,
		OID:       OID,
		Component: "interface",
		Operation: operation,
		Details: audit.Details{
			ID:          "LAN",
			Name:        stringify(l.Name, ""),
			Field:       field,
			Description: description,
		},
	}

	if dbc != nil {
		dbc.Write(record)
	}
}
