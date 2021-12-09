package interfaces

import (
	"reflect"
	"testing"
	"time"

	core "github.com/uhppoted/uhppote-core/types"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/types"
)

func TestLANAsObjects(t *testing.T) {
	created = types.DateTime(time.Date(2021, time.February, 28, 12, 34, 56, 0, time.Local))
	bind, _ := core.ResolveBindAddr("192.168.1.101")
	broadcast, _ := core.ResolveBroadcastAddr("192.168.1.102")
	listen, _ := core.ResolveListenAddr("192.168.1.103:54321")

	l := LANx{
		OID:              "0.1.3",
		Name:             "Le LAN",
		BindAddress:      *bind,
		BroadcastAddress: *broadcast,
		ListenAddress:    *listen,

		created: created,
	}

	expected := []interface{}{
		catalog.Object{OID: "0.1.3", Value: ""},
		catalog.Object{OID: "0.1.3.0.0", Value: types.StatusOk},
		catalog.Object{OID: "0.1.3.0.1", Value: created},
		catalog.Object{OID: "0.1.3.0.2", Value: (*types.DateTime)(nil)},
		catalog.Object{OID: "0.1.3.0.4", Value: "LAN"},
		catalog.Object{OID: "0.1.3.1", Value: "Le LAN"},
		catalog.Object{OID: "0.1.3.3.1", Value: *bind},
		catalog.Object{OID: "0.1.3.3.2", Value: *broadcast},
		catalog.Object{OID: "0.1.3.3.3", Value: *listen},
	}

	objects := l.AsObjects()

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Incorrect return from AsObjects:\n   expected:%#v\n   got:     %#v", expected, objects)
	}
}

func TestLANAsObjectsWithDeleted(t *testing.T) {
	created = types.DateTime(time.Date(2021, time.February, 28, 12, 34, 56, 0, time.Local))
	deleted := types.DateTimePtrNow()
	bind, _ := core.ResolveBindAddr("192.168.1.101")
	broadcast, _ := core.ResolveBroadcastAddr("192.168.1.102")
	listen, _ := core.ResolveListenAddr("192.168.1.103:54321")

	l := LANx{
		OID:              "0.1.3",
		Name:             "Le LAN",
		BindAddress:      *bind,
		BroadcastAddress: *broadcast,
		ListenAddress:    *listen,

		created: created,
		deleted: deleted,
	}

	expected := []interface{}{
		catalog.Object{
			OID:   "0.1.3.0.2",
			Value: deleted,
		},
	}

	objects := l.AsObjects()

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Incorrect return from AsObjects:\n   expected:%#v\n   got:     %#v", expected, objects)
	}
}
