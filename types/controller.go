package types

import (
	"net"
	"time"

	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
)

type IController interface {
	OID() schema.OID
	Name() string
	ID() uint32
	EndPoint() *net.UDPAddr
	TimeZone() *time.Location
	Door(uint8) (schema.OID, bool)

	DateTimeOk() bool
}
