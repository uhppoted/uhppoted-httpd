package system

import (
	"time"

	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	"github.com/uhppoted/uhppoted-httpd/types"
	"github.com/uhppoted/uhppoted-lib/uhppoted"
)

func Events(uid, role string, start, count int) []schema.Object {
	sys.RLock()
	defer sys.RUnlock()

	auth := auth.NewAuthorizator(uid, role)
	objects := sys.events.AsObjects(start, count, auth)

	return objects
}

func AppendEvents(list types.EventsList) {
	deviceID := list.DeviceID
	recent := list.Events

	l := func(e uhppoted.Event) (string, string, string) {
		timestamp := time.Time(e.Timestamp)
		deviceID := e.DeviceID
		doorID := e.Door
		cardID := e.CardNumber

		device := sys.history.LookupController(timestamp, deviceID)
		door := sys.history.LookupDoor(timestamp, deviceID, doorID)
		name := sys.history.LookupCard(timestamp, cardID)

		return device, door, name
	}

	sys.events.Received(deviceID, recent, l)

	if len(recent) > 0 {
		if err := save(TagEvents, &sys.events); err != nil {
			warnf("events", "%v", err)
		}
	}
}
