package memdb

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/uhppoted/uhppoted-httpd/audit"
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/cards"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/types"
)

var hagrid = cardholder("C01", "Hagrid", 6514231)
var dobby = cardholder("C02", "Dobby", 1234567, "G05")

type stub struct {
	write func(e audit.LogEntry)
}

func (x *stub) UID() string {
	return "stub"
}

func (x *stub) CanUpdateInterface(lan auth.Operant, field string, value interface{}) error {
	return fmt.Errorf("not authorised")
}

func (x *stub) CanAddController(controller auth.Operant) error {
	return fmt.Errorf("not authorised")
}

func (x *stub) CanUpdateController(controller auth.Operant, field string, value interface{}) error {
	return fmt.Errorf("not authorised")
}

func (x *stub) CanDeleteController(controller auth.Operant) error {
	return fmt.Errorf("not authorised")
}

func (x *stub) CanAddCardHolder(cardHolder auth.Operant) error {
	return fmt.Errorf("not authorised")
}

func (x *stub) CanUpdateCard(door auth.Operant, field string, value interface{}) error {
	return fmt.Errorf("not authorised")
}

func (x *stub) CanUpdateCardHolder(before, after auth.Operant) error {
	return fmt.Errorf("not authorised")
}

func (x *stub) CanDeleteCardHolder(cardHolder auth.Operant) error {
	return fmt.Errorf("not authorised")
}

func (x *stub) CanAddDoor(door auth.Operant) error {
	return fmt.Errorf("not authorised")
}

func (x *stub) CanUpdateDoor(door auth.Operant, field string, value interface{}) error {
	return fmt.Errorf("not authorised")
}

func (x *stub) CanDeleteDoor(door auth.Operant) error {
	return fmt.Errorf("not authorised")
}

func (x *stub) Write(e audit.LogEntry) {
	x.write(e)
}

func date(s string) *types.Date {
	date, _ := time.ParseInLocation("2006-01-02", s, time.Local)
	d := types.Date(date)

	return &d
}

func dbx(cardholders ...cards.CardHolder) *fdb {
	p := fdb{
		data: data{
			Tables: tables{
				Groups: types.Groups{
					"G05": group("G05"),
				},
				CardHolders: cards.CardHolders{},
			},
		},
	}

	for i, _ := range cardholders {
		c := cardholders[i].Clone()
		p.data.Tables.CardHolders[c.OID] = c
	}

	return &p
}

func group(id string) types.Group {
	return types.Group{
		OID:  id,
		Name: "",
	}
}

func cardholder(id, name string, card uint32, groups ...string) cards.CardHolder {
	n := types.Name(name)
	c := types.Card(card)

	cardholder := cards.CardHolder{
		OID:    catalog.OID(id),
		Name:   &n,
		Card:   &c,
		From:   date("2021-01-02"),
		To:     date("2021-12-30"),
		Groups: map[string]bool{},
	}

	for _, g := range groups {
		cardholder.Groups[g] = true
	}

	return cardholder
}

func compare(got, expected interface{}, t *testing.T) {
	p, _ := json.Marshal(got)
	q, _ := json.Marshal(expected)

	if string(p) != string(q) {
		t.Errorf("'got' does not match 'expected'\nexpected:%s\ngot:     %s", string(q), string(p))
	}
}

func compareDB(db, expected *fdb, t *testing.T) {
	compare(db.data.Tables, expected.data.Tables, t)
}
