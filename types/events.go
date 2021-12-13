package types

import (
	"github.com/uhppoted/uhppoted-lib/uhppoted"
)

type EventsList struct {
	DeviceID uint32
	Events   []uhppoted.Event
}
