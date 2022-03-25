package controllers

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	core "github.com/uhppoted/uhppote-core/types"

	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/impl"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	"github.com/uhppoted/uhppoted-httpd/types"
)

func TestControllerAsObjects(t *testing.T) {
	catalog.Init(memdb.Catalog())

	created = types.Timestamp(time.Date(2021, time.February, 28, 12, 34, 56, 0, time.Local))
	name := "Le Thing"
	deviceID := uint32(12345678)
	address, _ := core.ResolveAddr("192.168.1.101")

	c := Controller{
		CatalogController: catalog.CatalogController{
			OID:      "0.2.3",
			DeviceID: deviceID,
		},
		name: name,
		IP:   address,
		Doors: map[uint8]schema.OID{
			1: "0.3.5",
			2: "0.3.7",
			3: "0.3.9",
			4: "0.3.11",
		},
		created: created,
	}

	expected := []schema.Object{
		{OID: "0.2.3", Value: ""},
		{OID: "0.2.3.0.0", Value: types.StatusUnknown},
		{OID: "0.2.3.0.1", Value: created},
		{OID: "0.2.3.0.2", Value: types.Timestamp{}},
		{OID: "0.2.3.1", Value: name},
		{OID: "0.2.3.2", Value: fmt.Sprintf("%v", deviceID)},
		{OID: "0.2.3.3.0", Value: types.StatusUnknown},
		{OID: "0.2.3.3.1", Value: fmt.Sprintf("%v", address)},
		{OID: "0.2.3.3.2", Value: fmt.Sprintf("%v", address)},
		{OID: "0.2.3.4.0", Value: types.StatusUnknown},
		{OID: "0.2.3.4.1", Value: ""},
		{OID: "0.2.3.4.2", Value: ""},
		{OID: "0.2.3.5.0", Value: types.StatusUnknown},
		{OID: "0.2.3.5.1", Value: ""},
		{OID: "0.2.3.6.0", Value: types.StatusUnknown},
		{OID: "0.2.3.6.1", Value: types.Uint32(0)},
		{OID: "0.2.3.6.2", Value: types.Uint32(0)},
		{OID: "0.2.3.6.3", Value: types.Uint32(0)},
		{OID: "0.2.3.7.1", Value: schema.OID("0.3.5")},
		{OID: "0.2.3.7.2", Value: schema.OID("0.3.7")},
		{OID: "0.2.3.7.3", Value: schema.OID("0.3.9")},
		{OID: "0.2.3.7.4", Value: schema.OID("0.3.11")},
	}

	objects := c.AsObjects(nil)

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Incorrect return from AsObjects:\n   expected:%#v\n   got:     %#v", expected, objects)
	}
}

func TestControllerAsObjectsWithDeleted(t *testing.T) {
	created = types.Timestamp(time.Date(2021, time.February, 28, 12, 34, 56, 0, time.Local))
	deleted := types.TimestampNow()
	name := "Le Thing"
	deviceID := uint32(12345678)
	address, _ := core.ResolveAddr("192.168.1.101")

	c := Controller{
		CatalogController: catalog.CatalogController{
			OID:      "0.2.3",
			DeviceID: deviceID,
		},
		name: name,
		IP:   address,
		Doors: map[uint8]schema.OID{
			1: "0.3.5",
			2: "0.3.7",
			3: "0.3.9",
			4: "0.3.11",
		},
		created: created,
		deleted: deleted,
	}

	expected := []schema.Object{
		{OID: "0.2.3.0.2", Value: deleted},
	}

	objects := c.AsObjects(nil)

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Incorrect return from AsObjects:\n   expected:%#v\n   got:     %#v", expected, objects)
	}
}

func TestControllerAsObjectsWithAuth(t *testing.T) {
	created = types.Timestamp(time.Date(2021, time.February, 28, 12, 34, 56, 0, time.Local))
	name := "Le Thing"
	deviceID := uint32(12345678)
	address, _ := core.ResolveAddr("192.168.1.101")

	c := Controller{
		CatalogController: catalog.CatalogController{
			OID:      "0.2.3",
			DeviceID: deviceID,
		},
		name: name,
		IP:   address,
		Doors: map[uint8]schema.OID{
			1: "0.3.5",
			2: "0.3.7",
			3: "0.3.9",
			4: "0.3.11",
		},
		created: created,
	}

	expected := []schema.Object{
		{OID: "0.2.3", Value: ""},
		{OID: "0.2.3.0.0", Value: types.StatusUnknown},
		{OID: "0.2.3.0.1", Value: created},
		{OID: "0.2.3.0.2", Value: types.Timestamp{}},
		{OID: "0.2.3.1", Value: name},
		// {OID: "0.2.3.2", Value: fmt.Sprintf("%v", deviceID)},
		{OID: "0.2.3.3.0", Value: types.StatusUnknown},
		{OID: "0.2.3.3.1", Value: fmt.Sprintf("%v", address)},
		{OID: "0.2.3.3.2", Value: fmt.Sprintf("%v", address)},
		{OID: "0.2.3.4.0", Value: types.StatusUnknown},
		{OID: "0.2.3.4.1", Value: ""},
		{OID: "0.2.3.4.2", Value: ""},
		{OID: "0.2.3.5.0", Value: types.StatusUnknown},
		{OID: "0.2.3.5.1", Value: ""},
		{OID: "0.2.3.6.0", Value: types.StatusUnknown},
		{OID: "0.2.3.6.1", Value: types.Uint32(0)},
		{OID: "0.2.3.6.2", Value: types.Uint32(0)},
		{OID: "0.2.3.6.3", Value: types.Uint32(0)},
		{OID: "0.2.3.7.1", Value: schema.OID("0.3.5")},
		{OID: "0.2.3.7.2", Value: schema.OID("0.3.7")},
		{OID: "0.2.3.7.3", Value: schema.OID("0.3.9")},
		{OID: "0.2.3.7.4", Value: schema.OID("0.3.11")},
	}

	auth := stub{
		canView: func(ruleset auth.RuleSet, object auth.Operant, field string, value interface{}) error {
			if strings.HasPrefix(field, "controller.device.ID") {
				return errors.New("test")
			}

			return nil
		},
	}

	objects := c.AsObjects(&auth)

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Incorrect return from AsObjects:\n   expected:%#v\n   got:     %#v", expected, objects)
	}
}

func TestControllerSet(t *testing.T) {
	expected := []schema.Object{
		{OID: "0.2.3", Value: ""},
		{OID: "0.2.3.1", Value: "Ze Kontroller"},
		{OID: "0.2.3.0.0", Value: types.StatusUnknown},
	}

	c := Controller{
		CatalogController: catalog.CatalogController{
			OID: "0.2.3",
		},
		name: "Le Controlleur",
	}

	objects, err := c.set(nil, "0.2.3.1", "Ze Kontroller", nil)
	if err != nil {
		t.Errorf("Unexpected error (%v)", err)
	}

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Invalid result\n   expected:%#v\n   got:     %#v", expected, objects)
	}

	if c.name != "Ze Kontroller" {
		t.Errorf("Controller name not updated - expected:%v, got:%v", "Ze Kontroller", c.name)
	}
}

func TestControllerSetWithDeleted(t *testing.T) {
	c := Controller{
		CatalogController: catalog.CatalogController{
			OID: "0.2.3",
		},
		name: "Le Controlleur",

		deleted: types.TimestampNow(),
	}

	expected := []schema.Object{
		schema.Object{OID: "0.2.3.0.2", Value: c.deleted},
	}

	objects, err := c.set(nil, "0.2.3.1", "Ze Kontroller", nil)
	if err == nil {
		t.Errorf("Expected error, got (%v)", err)
	}

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Invalid result\n   expected:%#v\n   got:     %#v", expected, objects)
	}

	if c.name != "Le Controlleur" {
		t.Errorf("Controller name unexpectedly updated - expected:%v, got:%v", "Le Controlleur", c.name)
	}
}
