package doors

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	core "codeberg.org/uhppoted/uhppoted-core/types"

	"codeberg.org/uhppoted/uhppoted-httpd/auth"
	"codeberg.org/uhppoted/uhppoted-httpd/system/catalog"
	"codeberg.org/uhppoted/uhppoted-httpd/system/catalog/impl"
	"codeberg.org/uhppoted/uhppoted-httpd/system/catalog/schema"
	"codeberg.org/uhppoted/uhppoted-httpd/system/db"
	"codeberg.org/uhppoted/uhppoted-httpd/types"
)

func TestDoorAsObjects(t *testing.T) {
	catalog.Init(memdb.NewCatalog())

	created = types.Timestamp(time.Date(2021, time.February, 28, 12, 34, 56, 0, time.Local))

	d := Door{
		OID:       "0.3.3",
		name:      "Le Door",
		delay:     7,
		mode:      core.NormallyOpen,
		keypad:    true,
		passcodes: []uint32{12345, 999999, 54321},
		firstcard: core.FirstCard{
			StartTime: core.MustParseHHmm("08:30"),
			EndTime:   core.MustParseHHmm("16:45"),
			Active:    core.NormallyOpen,
			Inactive:  core.FirstCardOnly,
			Weekdays: core.Weekdays{
				time.Monday:   true,
				time.Tuesday:  true,
				time.Thursday: true,
				time.Sunday:   true,
			},
		},
		created: created,
	}

	expected := []schema.Object{
		{OID: "0.3.3", Value: ""},
		{OID: "0.3.3.0.0", Value: types.StatusOk},
		{OID: "0.3.3.0.1", Value: created},
		{OID: "0.3.3.0.2", Value: types.Timestamp{}},
		{OID: "0.3.3.1", Value: "Le Door"},
		{OID: "0.3.3.2", Value: types.Uint8(0)},
		{OID: "0.3.3.2.1", Value: types.StatusUnknown},
		{OID: "0.3.3.2.2", Value: uint8(7)},
		{OID: "0.3.3.2.3", Value: ""},
		{OID: "0.3.3.3", Value: core.ControlState(0)},
		{OID: "0.3.3.3.1", Value: types.StatusUnknown},
		{OID: "0.3.3.3.2", Value: core.NormallyOpen},
		{OID: "0.3.3.3.3", Value: ""},
		{OID: "0.3.3.4", Value: true},
		{OID: "0.3.3.5", Value: "******"},
		{OID: "0.3.3.6.1", Value: "08:30"},
		{OID: "0.3.3.6.2", Value: "16:45"},
		{OID: "0.3.3.6.3", Value: "normally open"},
		{OID: "0.3.3.6.4", Value: "firstcard only"},
		{OID: "0.3.3.6.5.1", Value: true},
		{OID: "0.3.3.6.5.2", Value: true},
		{OID: "0.3.3.6.5.3", Value: false},
		{OID: "0.3.3.6.5.4", Value: true},
		{OID: "0.3.3.6.5.5", Value: false},
		{OID: "0.3.3.6.5.6", Value: false},
		{OID: "0.3.3.6.5.7", Value: true},
	}

	objects := d.AsObjects(nil)

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Incorrect return from AsObjects:\n   expected:%#v\n   got:     %#v", expected, objects)
	}
}

func TestDoorAsObjectsWithDeleted(t *testing.T) {
	created = types.Timestamp(time.Date(2021, time.February, 28, 12, 34, 56, 0, time.Local))
	deleted := types.TimestampNow()

	d := Door{
		OID:       "0.3.3",
		name:      "Le Door",
		delay:     7,
		mode:      core.NormallyOpen,
		keypad:    true,
		passcodes: []uint32{12345, 999999, 54321},
		created:   created,
		deleted:   deleted,
	}

	expected := []schema.Object{
		{OID: "0.3.3.0.2", Value: deleted},
	}

	objects := d.AsObjects(nil)

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Incorrect return from AsObjects:\n   expected:%#v\n   got:     %#v", expected, objects)
	}
}

func TestDoorAsObjectsWithAuth(t *testing.T) {
	created = types.Timestamp(time.Date(2021, time.February, 28, 12, 34, 56, 0, time.Local))

	d := Door{
		OID:       "0.3.3",
		name:      "Le Door",
		delay:     7,
		mode:      core.NormallyOpen,
		keypad:    true,
		passcodes: []uint32{12345, 999999, 54321},
		firstcard: core.FirstCard{
			StartTime: core.MustParseHHmm("08:30"),
			EndTime:   core.MustParseHHmm("16:45"),
			Active:    core.NormallyOpen,
			Inactive:  core.FirstCardOnly,
			Weekdays: core.Weekdays{
				time.Monday:   true,
				time.Tuesday:  true,
				time.Thursday: true,
				time.Sunday:   true,
			},
		},
		created: created,
	}

	expected := []schema.Object{
		{OID: "0.3.3", Value: ""},
		{OID: "0.3.3.0.0", Value: types.StatusOk},
		{OID: "0.3.3.0.1", Value: created},
		{OID: "0.3.3.0.2", Value: types.Timestamp{}},
		{OID: "0.3.3.1", Value: "Le Door"},
		{OID: "0.3.3.3", Value: core.ControlState(0)},
		{OID: "0.3.3.3.1", Value: types.StatusUnknown},
		{OID: "0.3.3.3.2", Value: core.NormallyOpen},
		{OID: "0.3.3.3.3", Value: ""},
		{OID: "0.3.3.4", Value: true},
		{OID: "0.3.3.5", Value: "******"},
		{OID: "0.3.3.6.1", Value: "08:30"},
		{OID: "0.3.3.6.2", Value: "16:45"},
		{OID: "0.3.3.6.3", Value: "normally open"},
		{OID: "0.3.3.6.4", Value: "firstcard only"},
		{OID: "0.3.3.6.5.1", Value: true},
		{OID: "0.3.3.6.5.2", Value: true},
		{OID: "0.3.3.6.5.3", Value: false},
		{OID: "0.3.3.6.5.4", Value: true},
		{OID: "0.3.3.6.5.5", Value: false},
		{OID: "0.3.3.6.5.6", Value: false},
		{OID: "0.3.3.6.5.7", Value: true},
	}

	a := auth.Authorizator{
		OpAuth: &stub{
			canView: func(ruleset auth.RuleSet, object auth.Operant, field string, value any) error {
				if strings.HasPrefix(field, "door.delay") {
					return errors.New("test")
				}

				return nil
			},
		},
	}

	objects := d.AsObjects(&a)

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Incorrect return from AsObjects:\n   expected:%#v\n   got:     %#v", expected, objects)
	}
}

func TestDoorSet(t *testing.T) {
	expected := []schema.Object{
		schema.Object{OID: "0.3.3", Value: ""},
		schema.Object{OID: "0.3.3.1", Value: "Eine Kleine Dooren"},
		schema.Object{OID: "0.3.3.0.0", Value: types.StatusOk},
	}

	d := Door{
		OID:   "0.3.3",
		name:  "Le Door",
		delay: 7,
		mode:  core.NormallyOpen,
	}

	objects, err := d.set(nil, "0.3.3.1", "Eine Kleine Dooren", db.DBC{})
	if err != nil {
		t.Errorf("Unexpected error (%v)", err)
	}

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Invalid result\n   expected:%#v\n   got:     %#v", expected, objects)
	}

	if d.name != "Eine Kleine Dooren" {
		t.Errorf("Door name not updated - expected:%v, got:%v", "Eine Kleine Dooren", d.name)
	}
}

func TestDoorSetFirstCardStartTime(t *testing.T) {
	expected := []schema.Object{
		schema.Object{OID: "0.3.3", Value: ""},
		schema.Object{OID: "0.3.3.6.1", Value: "08:35"},
		schema.Object{OID: "0.3.3.0.0", Value: types.StatusOk},
	}

	d := Door{
		OID:   "0.3.3",
		name:  "Le Door",
		delay: 5,
		mode:  core.NormallyOpen,
	}

	objects, err := d.set(nil, "0.3.3.6.1", "08:35", db.DBC{})
	if err != nil {
		t.Errorf("Unexpected error (%v)", err)
	}

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Invalid result\n   expected:%#v\n   got:     %#v", expected, objects)
	}

	if d.firstcard.StartTime != core.MustParseHHmm("08:35") {
		t.Errorf("Door first card start time not updated - expected:%v, got:%v", "08:35", d.firstcard.StartTime)
	}
}

func TestDoorSetFirstCardEndTime(t *testing.T) {
	expected := []schema.Object{
		schema.Object{OID: "0.3.3", Value: ""},
		schema.Object{OID: "0.3.3.6.2", Value: "16:57"},
		schema.Object{OID: "0.3.3.0.0", Value: types.StatusOk},
	}

	d := Door{
		OID:   "0.3.3",
		name:  "Le Door",
		delay: 5,
		mode:  core.NormallyOpen,
	}

	objects, err := d.set(nil, "0.3.3.6.2", "16:57", db.DBC{})
	if err != nil {
		t.Errorf("Unexpected error (%v)", err)
	}

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Invalid result\n   expected:%#v\n   got:     %#v", expected, objects)
	}

	if d.firstcard.EndTime != core.MustParseHHmm("16:57") {
		t.Errorf("Door first card end time not updated - expected:%v, got:%v", "16:57", d.firstcard.EndTime)
	}
}

func TestDoorSetFirstCardActiveMode(t *testing.T) {
	expected := []schema.Object{
		schema.Object{OID: "0.3.3", Value: ""},
		schema.Object{OID: "0.3.3.6.3", Value: "normally closed"},
		schema.Object{OID: "0.3.3.0.0", Value: types.StatusOk},
	}

	d := Door{
		OID:   "0.3.3",
		name:  "Le Door",
		delay: 5,
		mode:  core.NormallyOpen,
	}

	objects, err := d.set(nil, "0.3.3.6.3", "normally closed", db.DBC{})
	if err != nil {
		t.Errorf("Unexpected error (%v)", err)
	}

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Invalid result\n   expected:%#v\n   got:     %#v", expected, objects)
	}

	if d.firstcard.Active != core.NormallyClosed {
		t.Errorf("Door first card active mode not updated - expected:%v, got:%v", core.NormallyClosed, d.firstcard.Active)
	}
}

func TestDoorSetFirstCardInvalidActiveMode(t *testing.T) {
	expected := []schema.Object{
		schema.Object{OID: "0.3.3", Value: ""},
		schema.Object{OID: "0.3.3.6.3", Value: "normally open"},
		schema.Object{OID: "0.3.3.0.0", Value: types.StatusOk},
	}

	d := Door{
		OID:   "0.3.3",
		name:  "Le Door",
		delay: 5,
		mode:  core.NormallyOpen,
		firstcard: core.FirstCard{
			Active: core.NormallyOpen,
		},
	}

	objects, err := d.set(nil, "0.3.3.6.3", "firstcard only", db.DBC{})
	if err != nil {
		t.Errorf("Unexpected error (%v)", err)
	}

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Invalid result\n   expected:%#v\n   got:     %#v", expected, objects)
	}

	if d.firstcard.Active != core.NormallyOpen {
		t.Errorf("Door first card active mode not updated - expected:%v, got:%v", core.NormallyOpen, d.firstcard.Active)
	}
}

func TestDoorSetFirstCardInactiveMode(t *testing.T) {
	expected := []schema.Object{
		schema.Object{OID: "0.3.3", Value: ""},
		schema.Object{OID: "0.3.3.6.4", Value: "firstcard only"},
		schema.Object{OID: "0.3.3.0.0", Value: types.StatusOk},
	}

	d := Door{
		OID:   "0.3.3",
		name:  "Le Door",
		delay: 5,
		mode:  core.NormallyOpen,
		firstcard: core.FirstCard{
			Active: core.NormallyOpen,
		},
	}

	objects, err := d.set(nil, "0.3.3.6.4", "firstcard only", db.DBC{})
	if err != nil {
		t.Errorf("Unexpected error (%v)", err)
	}

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Invalid result\n   expected:%#v\n   got:     %#v", expected, objects)
	}

	if d.firstcard.Inactive != core.FirstCardOnly {
		t.Errorf("Door first card inactive mode not updated - expected:%v, got:%v", core.FirstCardOnly, d.firstcard.Inactive)
	}
}

func TestDoorSetFirstCardInvalidInactiveMode(t *testing.T) {
	expected := []schema.Object{
		schema.Object{OID: "0.3.3", Value: ""},
		schema.Object{OID: "0.3.3.6.4", Value: "firstcard only"},
		schema.Object{OID: "0.3.3.0.0", Value: types.StatusOk},
	}

	d := Door{
		OID:   "0.3.3",
		name:  "Le Door",
		delay: 5,
		mode:  core.NormallyOpen,
		firstcard: core.FirstCard{
			Inactive: core.FirstCardOnly,
		},
	}

	objects, err := d.set(nil, "0.3.3.6.4", "qwertyuiop", db.DBC{})
	if err != nil {
		t.Errorf("Unexpected error (%v)", err)
	}

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Invalid result\n   expected:%#v\n   got:     %#v", expected, objects)
	}

	if d.firstcard.Inactive != core.FirstCardOnly {
		t.Errorf("Door first card active mode unexpectedly updated - expected:%v, got:%v", core.FirstCardOnly, d.firstcard.Inactive)
	}
}

func TestDoorSetFirstCardMonday(t *testing.T) {
	expected := []schema.Object{
		schema.Object{OID: "0.3.3", Value: ""},
		schema.Object{OID: "0.3.3.6.5.1", Value: "true"},
		schema.Object{OID: "0.3.3.0.0", Value: types.StatusOk},
	}

	d := Door{
		OID:   "0.3.3",
		name:  "Le Door",
		delay: 5,
		mode:  core.NormallyOpen,
		firstcard: core.FirstCard{
			Weekdays: core.Weekdays{
				time.Monday:    false,
				time.Tuesday:   false,
				time.Wednesday: false,
				time.Thursday:  false,
				time.Friday:    false,
				time.Saturday:  false,
				time.Sunday:    false,
			},
		},
	}

	objects, err := d.set(nil, "0.3.3.6.5.1", "true", db.DBC{})
	if err != nil {
		t.Errorf("Unexpected error (%v)", err)
	}

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Invalid result\n   expected:%#v\n   got:     %#v", expected, objects)
	}

	if !d.firstcard.Weekdays[time.Monday] {
		t.Errorf("Door first card weekdays.monday mode not updated - expected:%v, got:%v", true, d.firstcard.Weekdays[time.Monday])
	}

	if d.firstcard.Weekdays[time.Tuesday] {
		t.Errorf("Door first card weekdays.tuesday mode unexpected updated - expected:%v, got:%v", false, d.firstcard.Weekdays[time.Tuesday])
	}

	if d.firstcard.Weekdays[time.Wednesday] {
		t.Errorf("Door first card weekdays.wednesday mode unexpected updated - expected:%v, got:%v", false, d.firstcard.Weekdays[time.Wednesday])
	}

	if d.firstcard.Weekdays[time.Thursday] {
		t.Errorf("Door first card weekdays.thursday mode unexpected updated - expected:%v, got:%v", false, d.firstcard.Weekdays[time.Thursday])
	}

	if d.firstcard.Weekdays[time.Friday] {
		t.Errorf("Door first card weekdays.friday mode unexpected updated - expected:%v, got:%v", false, d.firstcard.Weekdays[time.Friday])
	}

	if d.firstcard.Weekdays[time.Saturday] {
		t.Errorf("Door first card weekdays.saturday mode unexpected updated - expected:%v, got:%v", false, d.firstcard.Weekdays[time.Saturday])
	}

	if d.firstcard.Weekdays[time.Sunday] {
		t.Errorf("Door first card weekdays.sunday mode unexpected updated - expected:%v, got:%v", false, d.firstcard.Weekdays[time.Sunday])
	}
}

func TestDoorSetFirstCardTuesday(t *testing.T) {
	expected := []schema.Object{
		schema.Object{OID: "0.3.3", Value: ""},
		schema.Object{OID: "0.3.3.6.5.2", Value: "true"},
		schema.Object{OID: "0.3.3.0.0", Value: types.StatusOk},
	}

	d := Door{
		OID:   "0.3.3",
		name:  "Le Door",
		delay: 5,
		mode:  core.NormallyOpen,
		firstcard: core.FirstCard{
			Weekdays: core.Weekdays{
				time.Monday:    false,
				time.Tuesday:   false,
				time.Wednesday: false,
				time.Thursday:  false,
				time.Friday:    false,
				time.Saturday:  false,
				time.Sunday:    false,
			},
		},
	}

	objects, err := d.set(nil, "0.3.3.6.5.2", "true", db.DBC{})
	if err != nil {
		t.Errorf("Unexpected error (%v)", err)
	}

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Invalid result\n   expected:%#v\n   got:     %#v", expected, objects)
	}

	if d.firstcard.Weekdays[time.Monday] {
		t.Errorf("Door first card weekdays.monday mode unexpectedly updated - expected:%v, got:%v", true, d.firstcard.Weekdays[time.Monday])
	}

	if !d.firstcard.Weekdays[time.Tuesday] {
		t.Errorf("Door first card weekdays.tuesday mode not updated - expected:%v, got:%v", false, d.firstcard.Weekdays[time.Tuesday])
	}

	if d.firstcard.Weekdays[time.Wednesday] {
		t.Errorf("Door first card weekdays.wednesday mode unexpectedly updated - expected:%v, got:%v", false, d.firstcard.Weekdays[time.Wednesday])
	}

	if d.firstcard.Weekdays[time.Thursday] {
		t.Errorf("Door first card weekdays.thursday mode unexpectedly updated - expected:%v, got:%v", false, d.firstcard.Weekdays[time.Thursday])
	}

	if d.firstcard.Weekdays[time.Friday] {
		t.Errorf("Door first card weekdays.friday mode unexpectedly updated - expected:%v, got:%v", false, d.firstcard.Weekdays[time.Friday])
	}

	if d.firstcard.Weekdays[time.Saturday] {
		t.Errorf("Door first card weekdays.saturday mode unexpectedly updated - expected:%v, got:%v", false, d.firstcard.Weekdays[time.Saturday])
	}

	if d.firstcard.Weekdays[time.Sunday] {
		t.Errorf("Door first card weekdays.sunday mode unexpectedly updated - expected:%v, got:%v", false, d.firstcard.Weekdays[time.Sunday])
	}
}

func TestDoorSetFirstCardWednesday(t *testing.T) {
	expected := []schema.Object{
		schema.Object{OID: "0.3.3", Value: ""},
		schema.Object{OID: "0.3.3.6.5.3", Value: "true"},
		schema.Object{OID: "0.3.3.0.0", Value: types.StatusOk},
	}

	d := Door{
		OID:   "0.3.3",
		name:  "Le Door",
		delay: 5,
		mode:  core.NormallyOpen,
		firstcard: core.FirstCard{
			Weekdays: core.Weekdays{
				time.Monday:    false,
				time.Tuesday:   false,
				time.Wednesday: false,
				time.Thursday:  false,
				time.Friday:    false,
				time.Saturday:  false,
				time.Sunday:    false,
			},
		},
	}

	objects, err := d.set(nil, "0.3.3.6.5.3", "true", db.DBC{})
	if err != nil {
		t.Errorf("Unexpected error (%v)", err)
	}

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Invalid result\n   expected:%#v\n   got:     %#v", expected, objects)
	}

	if d.firstcard.Weekdays[time.Monday] {
		t.Errorf("Door first card weekdays.monday mode unexpectedly updated - expected:%v, got:%v", true, d.firstcard.Weekdays[time.Monday])
	}

	if d.firstcard.Weekdays[time.Tuesday] {
		t.Errorf("Door first card weekdays.tuesday mode unexpectedly updated - expected:%v, got:%v", false, d.firstcard.Weekdays[time.Tuesday])
	}

	if !d.firstcard.Weekdays[time.Wednesday] {
		t.Errorf("Door first card weekdays.wednesday mode not updated - expected:%v, got:%v", false, d.firstcard.Weekdays[time.Wednesday])
	}

	if d.firstcard.Weekdays[time.Thursday] {
		t.Errorf("Door first card weekdays.thursday mode unexpectedly updated - expected:%v, got:%v", false, d.firstcard.Weekdays[time.Thursday])
	}

	if d.firstcard.Weekdays[time.Friday] {
		t.Errorf("Door first card weekdays.friday mode unexpectedly updated - expected:%v, got:%v", false, d.firstcard.Weekdays[time.Friday])
	}

	if d.firstcard.Weekdays[time.Saturday] {
		t.Errorf("Door first card weekdays.saturday mode unexpectedly updated - expected:%v, got:%v", false, d.firstcard.Weekdays[time.Saturday])
	}

	if d.firstcard.Weekdays[time.Sunday] {
		t.Errorf("Door first card weekdays.sunday mode unexpectedly updated - expected:%v, got:%v", false, d.firstcard.Weekdays[time.Sunday])
	}
}

func TestDoorSetFirstCardThursday(t *testing.T) {
	expected := []schema.Object{
		schema.Object{OID: "0.3.3", Value: ""},
		schema.Object{OID: "0.3.3.6.5.4", Value: "true"},
		schema.Object{OID: "0.3.3.0.0", Value: types.StatusOk},
	}

	d := Door{
		OID:   "0.3.3",
		name:  "Le Door",
		delay: 5,
		mode:  core.NormallyOpen,
		firstcard: core.FirstCard{
			Weekdays: core.Weekdays{
				time.Monday:    false,
				time.Tuesday:   false,
				time.Wednesday: false,
				time.Thursday:  false,
				time.Friday:    false,
				time.Saturday:  false,
				time.Sunday:    false,
			},
		},
	}

	objects, err := d.set(nil, "0.3.3.6.5.4", "true", db.DBC{})
	if err != nil {
		t.Errorf("Unexpected error (%v)", err)
	}

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Invalid result\n   expected:%#v\n   got:     %#v", expected, objects)
	}

	if d.firstcard.Weekdays[time.Monday] {
		t.Errorf("Door first card weekdays.monday mode unexpectedly updated - expected:%v, got:%v", true, d.firstcard.Weekdays[time.Monday])
	}

	if d.firstcard.Weekdays[time.Tuesday] {
		t.Errorf("Door first card weekdays.tuesday mode unexpectedly updated - expected:%v, got:%v", false, d.firstcard.Weekdays[time.Tuesday])
	}

	if d.firstcard.Weekdays[time.Wednesday] {
		t.Errorf("Door first card weekdays.wednesday mode unexpectedly updated - expected:%v, got:%v", false, d.firstcard.Weekdays[time.Wednesday])
	}

	if !d.firstcard.Weekdays[time.Thursday] {
		t.Errorf("Door first card weekdays.thursday mode not updated - expected:%v, got:%v", false, d.firstcard.Weekdays[time.Thursday])
	}

	if d.firstcard.Weekdays[time.Friday] {
		t.Errorf("Door first card weekdays.friday mode unexpectedly updated - expected:%v, got:%v", false, d.firstcard.Weekdays[time.Friday])
	}

	if d.firstcard.Weekdays[time.Saturday] {
		t.Errorf("Door first card weekdays.saturday mode unexpectedly updated - expected:%v, got:%v", false, d.firstcard.Weekdays[time.Saturday])
	}

	if d.firstcard.Weekdays[time.Sunday] {
		t.Errorf("Door first card weekdays.sunday mode unexpectedly updated - expected:%v, got:%v", false, d.firstcard.Weekdays[time.Sunday])
	}
}

func TestDoorSetFirstCardFriday(t *testing.T) {
	expected := []schema.Object{
		schema.Object{OID: "0.3.3", Value: ""},
		schema.Object{OID: "0.3.3.6.5.5", Value: "true"},
		schema.Object{OID: "0.3.3.0.0", Value: types.StatusOk},
	}

	d := Door{
		OID:   "0.3.3",
		name:  "Le Door",
		delay: 5,
		mode:  core.NormallyOpen,
		firstcard: core.FirstCard{
			Weekdays: core.Weekdays{
				time.Monday:    false,
				time.Tuesday:   false,
				time.Wednesday: false,
				time.Thursday:  false,
				time.Friday:    false,
				time.Saturday:  false,
				time.Sunday:    false,
			},
		},
	}

	objects, err := d.set(nil, "0.3.3.6.5.5", "true", db.DBC{})
	if err != nil {
		t.Errorf("Unexpected error (%v)", err)
	}

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Invalid result\n   expected:%#v\n   got:     %#v", expected, objects)
	}

	if d.firstcard.Weekdays[time.Monday] {
		t.Errorf("Door first card weekdays.monday mode unexpectedly updated - expected:%v, got:%v", true, d.firstcard.Weekdays[time.Monday])
	}

	if d.firstcard.Weekdays[time.Tuesday] {
		t.Errorf("Door first card weekdays.tuesday mode unexpectedly updated - expected:%v, got:%v", false, d.firstcard.Weekdays[time.Tuesday])
	}

	if d.firstcard.Weekdays[time.Wednesday] {
		t.Errorf("Door first card weekdays.wednesday mode unexpectedly updated - expected:%v, got:%v", false, d.firstcard.Weekdays[time.Wednesday])
	}

	if d.firstcard.Weekdays[time.Thursday] {
		t.Errorf("Door first card weekdays.thursday mode unexpectedly updated - expected:%v, got:%v", false, d.firstcard.Weekdays[time.Thursday])
	}

	if !d.firstcard.Weekdays[time.Friday] {
		t.Errorf("Door first card weekdays.friday mode not updated - expected:%v, got:%v", false, d.firstcard.Weekdays[time.Friday])
	}

	if d.firstcard.Weekdays[time.Saturday] {
		t.Errorf("Door first card weekdays.saturday mode unexpectedly updated - expected:%v, got:%v", false, d.firstcard.Weekdays[time.Saturday])
	}

	if d.firstcard.Weekdays[time.Sunday] {
		t.Errorf("Door first card weekdays.sunday mode unexpectedly updated - expected:%v, got:%v", false, d.firstcard.Weekdays[time.Sunday])
	}
}

func TestDoorSetFirstCardSaturday(t *testing.T) {
	expected := []schema.Object{
		schema.Object{OID: "0.3.3", Value: ""},
		schema.Object{OID: "0.3.3.6.5.6", Value: "true"},
		schema.Object{OID: "0.3.3.0.0", Value: types.StatusOk},
	}

	d := Door{
		OID:   "0.3.3",
		name:  "Le Door",
		delay: 5,
		mode:  core.NormallyOpen,
		firstcard: core.FirstCard{
			Weekdays: core.Weekdays{
				time.Monday:    false,
				time.Tuesday:   false,
				time.Wednesday: false,
				time.Thursday:  false,
				time.Friday:    false,
				time.Saturday:  false,
				time.Sunday:    false,
			},
		},
	}

	objects, err := d.set(nil, "0.3.3.6.5.6", "true", db.DBC{})
	if err != nil {
		t.Errorf("Unexpected error (%v)", err)
	}

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Invalid result\n   expected:%#v\n   got:     %#v", expected, objects)
	}

	if d.firstcard.Weekdays[time.Monday] {
		t.Errorf("Door first card weekdays.monday mode unexpectedly updated - expected:%v, got:%v", true, d.firstcard.Weekdays[time.Monday])
	}

	if d.firstcard.Weekdays[time.Tuesday] {
		t.Errorf("Door first card weekdays.tuesday mode unexpectedly updated - expected:%v, got:%v", false, d.firstcard.Weekdays[time.Tuesday])
	}

	if d.firstcard.Weekdays[time.Wednesday] {
		t.Errorf("Door first card weekdays.wednesday mode unexpectedly updated - expected:%v, got:%v", false, d.firstcard.Weekdays[time.Wednesday])
	}

	if d.firstcard.Weekdays[time.Thursday] {
		t.Errorf("Door first card weekdays.thursday mode unexpectedly updated - expected:%v, got:%v", false, d.firstcard.Weekdays[time.Thursday])
	}

	if d.firstcard.Weekdays[time.Friday] {
		t.Errorf("Door first card weekdays.friday mode unexpectedly updated - expected:%v, got:%v", false, d.firstcard.Weekdays[time.Friday])
	}

	if !d.firstcard.Weekdays[time.Saturday] {
		t.Errorf("Door first card weekdays.saturday mode not updated - expected:%v, got:%v", false, d.firstcard.Weekdays[time.Saturday])
	}

	if d.firstcard.Weekdays[time.Sunday] {
		t.Errorf("Door first card weekdays.sunday mode unexpectedly updated - expected:%v, got:%v", false, d.firstcard.Weekdays[time.Sunday])
	}
}

func TestDoorSetFirstCardSunday(t *testing.T) {
	expected := []schema.Object{
		schema.Object{OID: "0.3.3", Value: ""},
		schema.Object{OID: "0.3.3.6.5.7", Value: "true"},
		schema.Object{OID: "0.3.3.0.0", Value: types.StatusOk},
	}

	d := Door{
		OID:   "0.3.3",
		name:  "Le Door",
		delay: 5,
		mode:  core.NormallyOpen,
		firstcard: core.FirstCard{
			Weekdays: core.Weekdays{
				time.Monday:    false,
				time.Tuesday:   false,
				time.Wednesday: false,
				time.Thursday:  false,
				time.Friday:    false,
				time.Saturday:  false,
				time.Sunday:    false,
			},
		},
	}

	objects, err := d.set(nil, "0.3.3.6.5.7", "true", db.DBC{})
	if err != nil {
		t.Errorf("Unexpected error (%v)", err)
	}

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Invalid result\n   expected:%#v\n   got:     %#v", expected, objects)
	}

	if d.firstcard.Weekdays[time.Monday] {
		t.Errorf("Door first card weekdays.monday mode unexpectedly updated - expected:%v, got:%v", true, d.firstcard.Weekdays[time.Monday])
	}

	if d.firstcard.Weekdays[time.Tuesday] {
		t.Errorf("Door first card weekdays.tuesday mode unexpectedly updated - expected:%v, got:%v", false, d.firstcard.Weekdays[time.Tuesday])
	}

	if d.firstcard.Weekdays[time.Wednesday] {
		t.Errorf("Door first card weekdays.wednesday mode unexpectedly updated - expected:%v, got:%v", false, d.firstcard.Weekdays[time.Wednesday])
	}

	if d.firstcard.Weekdays[time.Thursday] {
		t.Errorf("Door first card weekdays.thursday mode unexpectedly updated - expected:%v, got:%v", false, d.firstcard.Weekdays[time.Thursday])
	}

	if d.firstcard.Weekdays[time.Friday] {
		t.Errorf("Door first card weekdays.friday mode unexpectedly updated - expected:%v, got:%v", false, d.firstcard.Weekdays[time.Friday])
	}

	if d.firstcard.Weekdays[time.Saturday] {
		t.Errorf("Door first card weekdays.saturday mode unexpectedly updated - expected:%v, got:%v", false, d.firstcard.Weekdays[time.Saturday])
	}

	if !d.firstcard.Weekdays[time.Sunday] {
		t.Errorf("Door first card weekdays.sunday mode not updated - expected:%v, got:%v", false, d.firstcard.Weekdays[time.Sunday])
	}
}

func TestDoorSetWithDeleted(t *testing.T) {
	d := Door{
		OID:   "0.3.3",
		name:  "Le Door",
		delay: 7,
		mode:  core.NormallyOpen,

		deleted: types.TimestampNow(),
	}

	expected := []schema.Object{
		schema.Object{OID: "0.3.3.0.2", Value: d.deleted},
	}

	objects, err := d.set(nil, "0.3.3.1", "Eine Kleine Dooren", db.DBC{})
	if err == nil {
		t.Errorf("Expected error, got (%v)", err)
	}

	if !reflect.DeepEqual(objects, expected) {
		t.Errorf("Invalid result\n   expected:%#v\n   got:     %#v", expected, objects)
	}

	if d.name != "Le Door" {
		t.Errorf("Door name unexpectedly updated - expected:%v, got:%v", "Le Door", d.name)
	}
}
