package interfaces

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	core "github.com/uhppoted/uhppote-core/types"

	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	"github.com/uhppoted/uhppoted-httpd/types"
)

func TestLANAsObjects(t *testing.T) {
	created = types.Timestamp(time.Date(2021, time.February, 28, 12, 34, 56, 0, time.Local))
	bind, _ := core.ResolveBindAddr("192.168.1.101")
	broadcast, _ := core.ResolveBroadcastAddr("192.168.1.102")
	listen, _ := core.ResolveListenAddr("192.168.1.103:54321")

	l := LAN{
		CatalogInterface: catalog.CatalogInterface{
			OID: "0.1.3",
		},
		Name:             "Le LAN",
		BindAddress:      *bind,
		BroadcastAddress: *broadcast,
		ListenAddress:    *listen,

		created: created,
	}

	expected := []schema.Object{
		{OID: "0.1.3", Value: ""},
		{OID: "0.1.3.0.0", Value: types.StatusOk},
		{OID: "0.1.3.0.1", Value: created},
		{OID: "0.1.3.0.2", Value: types.Timestamp{}},
		{OID: "0.1.3.0.4", Value: "LAN"},
		{OID: "0.1.3.1", Value: "Le LAN"},
		{OID: "0.1.3.3.1", Value: *bind},
		{OID: "0.1.3.3.2", Value: *broadcast},
		{OID: "0.1.3.3.3", Value: *listen},
	}

	objects := l.AsObjects(nil)

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Incorrect return from AsObjects:\n   expected:%#v\n   got:     %#v", expected, objects)
	}
}

func TestLANAsObjectsWithDeleted(t *testing.T) {
	created = types.Timestamp(time.Date(2021, time.February, 28, 12, 34, 56, 0, time.Local))
	deleted := types.TimestampNow()
	bind, _ := core.ResolveBindAddr("192.168.1.101")
	broadcast, _ := core.ResolveBroadcastAddr("192.168.1.102")
	listen, _ := core.ResolveListenAddr("192.168.1.103:54321")

	l := LAN{
		CatalogInterface: catalog.CatalogInterface{
			OID: "0.1.3",
		},
		Name:             "Le LAN",
		BindAddress:      *bind,
		BroadcastAddress: *broadcast,
		ListenAddress:    *listen,

		created: created,
		deleted: deleted,
	}

	expected := []schema.Object{
		{OID: "0.1.3.0.2", Value: deleted},
	}

	objects := l.AsObjects(nil)

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Incorrect return from AsObjects:\n   expected:%#v\n   got:     %#v", expected, objects)
	}
}

func TestLANAsObjectsWithAuth(t *testing.T) {
	created = types.Timestamp(time.Date(2021, time.February, 28, 12, 34, 56, 0, time.Local))
	bind, _ := core.ResolveBindAddr("192.168.1.101")
	broadcast, _ := core.ResolveBroadcastAddr("192.168.1.102")
	listen, _ := core.ResolveListenAddr("192.168.1.103:54321")

	l := LAN{
		CatalogInterface: catalog.CatalogInterface{
			OID: "0.1.3",
		},
		Name:             "Le LAN",
		BindAddress:      *bind,
		BroadcastAddress: *broadcast,
		ListenAddress:    *listen,

		created: created,
	}

	expected := []schema.Object{
		schema.Object{OID: "0.1.3", Value: ""},
		schema.Object{OID: "0.1.3.0.0", Value: types.StatusOk},
		schema.Object{OID: "0.1.3.0.1", Value: created},
		schema.Object{OID: "0.1.3.0.2", Value: types.Timestamp{}},
		schema.Object{OID: "0.1.3.0.4", Value: "LAN"},
		schema.Object{OID: "0.1.3.1", Value: "Le LAN"},
		schema.Object{OID: "0.1.3.3.1", Value: *bind},
		schema.Object{OID: "0.1.3.3.2", Value: *broadcast},
		//		schema.Object{OID: "0.1.3.3.3", Value: *listen},
	}

	auth := stub{
		canView: func(ruleset auth.RuleSet, object auth.Operant, field string, value interface{}) error {
			if strings.HasPrefix(field, "LAN.address.listen") {
				return errors.New("test")
			}

			return nil
		},
	}

	objects := l.AsObjects(&auth)

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Incorrect return from AsObjects:\n   expected:%#v\n   got:     %#v", expected, objects)
	}
}

func TestLANSet(t *testing.T) {
	expected := []schema.Object{
		schema.Object{OID: "0.1.3", Value: ""},
		schema.Object{OID: "0.1.3.1", Value: "Ze LAN"},
		schema.Object{OID: "0.1.3.0.0", Value: types.StatusOk},
	}

	l := LAN{
		CatalogInterface: catalog.CatalogInterface{
			OID: "0.1.3",
		},
		Name: "Le LAN",
	}

	objects, err := l.set(nil, "0.1.3.1", "Ze LAN", nil)
	if err != nil {
		t.Errorf("Unexpected error (%v)", err)
	}

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Invalid result\n   expected:%#v\n   got:     %#v", expected, objects)
	}

	if l.Name != "Ze LAN" {
		t.Errorf("LAN name not updated - expected:%v, got:%v", "Ze LAN", l.Name)
	}
}

func TestLANSetWithDeleted(t *testing.T) {
	l := LAN{
		CatalogInterface: catalog.CatalogInterface{
			OID: "0.1.3",
		},
		Name: "Le LAN",

		deleted: types.TimestampNow(),
	}

	expected := []schema.Object{
		schema.Object{OID: "0.1.3.0.2", Value: l.deleted},
	}

	objects, err := l.set(nil, "0.1.3.1", "Ze LAN", nil)
	if err == nil {
		t.Errorf("Expected error, got (%v)", err)
	}

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Invalid result\n   expected:%#v\n   got:     %#v", expected, objects)
	}

	if l.Name != "Le LAN" {
		t.Errorf("LAN name unexpectedly updated - expected:%v, got:%v", "Le LAN", l.Name)
	}
}

func TestLANSetWhileDeleting(t *testing.T) {
	l := LAN{
		CatalogInterface: catalog.CatalogInterface{
			OID: "0.1.3",
		},
		Name: "Le LAN",

		deleted:  types.TimestampNow(),
		deleting: true,
	}

	expected := []schema.Object{}

	objects, err := l.set(nil, "0.1.3.1", "Ze LAN", nil)
	if err != nil {
		t.Errorf("Unexpected error, got (%v)", err)
	}

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Invalid result\n   expected:%#v\n   got:     %#v", expected, objects)
	}

	if l.Name != "Le LAN" {
		t.Errorf("LAN name unexpectedly updated - expected:%v, got:%v", "Le LAN", l.Name)
	}
}
