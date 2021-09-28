package events

import (
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/types"
	"github.com/uhppoted/uhppoted-lib/uhppoted"
)

type Event struct {
	OID       catalog.OID    `json:"OID"`
	Timestamp types.DateTime `json:"timestamp"`
}

const EventTimestamp = catalog.EventTimestamp

// type event struct {
// 	device     uint32
// 	index      uint32
// 	eventType  uint8
// 	granted    bool
// 	door       uint8
// 	direction  uint8
// 	cardnumber uint32
// 	timestamp  types.DateTime
// 	reason     uint8
// }

func NewEvent(oid catalog.OID, e uhppoted.Event) Event {
	return Event{
		OID:       oid,
		Timestamp: types.DateTime(e.Timestamp),
	}
	//			events.cache = append(events.cache, event{
	//				device:     id,
	//				index:      e.Index,
	//				eventType:  e.Type,
	//				granted:    e.Granted,
	//				door:       e.Door,
	//				direction:  e.Direction,
	//				cardnumber: e.CardNumber,
	//				reason:     e.Reason,
	//			})

}

func (e Event) IsValid() bool {
	return true
}

func (e Event) IsDeleted() bool {
	return false
}

func (e *Event) AsObjects() []interface{} {
	objects := []interface{}{
		catalog.NewObject(e.OID, types.StatusOk),
		catalog.NewObject2(e.OID, EventTimestamp, e.Timestamp),
	}

	return objects
}

func (e Event) clone() Event {
	event := Event{
		OID:       e.OID,
		Timestamp: e.Timestamp,
	}

	return event
}

func (e *Event) set(auth auth.OpAuth, oid catalog.OID, value string) ([]interface{}, error) {
	objects := []interface{}{}

	return objects, nil
}
