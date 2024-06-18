package types

import (
	"time"

	"github.com/uhppoted/uhppote-core/types"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
)

type IController interface {
	OID() schema.OID
	Name() string
	ID() uint32
	EndPoint() types.ControllerAddr
	TimeZone() *time.Location
	Protocol() string
	Door(uint8) (schema.OID, bool)

	DateTimeOk() bool
}
