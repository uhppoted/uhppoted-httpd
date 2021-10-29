package cards

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/uhppoted/uhppoted-httpd/audit"
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/types"
)

var hagrid = makeCard("0.3.1", "Hagrid", 6514231)
var dobby = makeCard("0.3.2", "Dobby", 1234567, "G05")

type stub struct {
	canUpdateCard func(auth.Operant, string, interface{}) error
	write         func(e audit.AuditRecord)
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

func (x *stub) CanAddCard(card auth.Operant) error {
	return fmt.Errorf("not authorised")
}

func (x *stub) CanUpdateCard(card auth.Operant, field string, value interface{}) error {
	if x.canUpdateCard != nil {
		return x.canUpdateCard(card, field, value)
	}

	return fmt.Errorf("not authorised")
}

func (x *stub) CanDeleteCard(card auth.Operant) error {
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

func (x *stub) CanAddGroup(group auth.Operant) error {
	return fmt.Errorf("not authorised")
}

func (x *stub) CanUpdateGroup(group auth.Operant, field string, value interface{}) error {
	return fmt.Errorf("not authorised")
}

func (x *stub) CanDeleteGroup(group auth.Operant) error {
	return fmt.Errorf("not authorised")
}

func (x *stub) Write(e audit.AuditRecord) {
	x.write(e)
}

func date(s string) *types.Date {
	date, _ := time.ParseInLocation("2006-01-02", s, time.Local)
	d := types.Date(date)

	return &d
}

func makeCards(list ...Card) *Cards {
	p := Cards{
		Cards: map[catalog.OID]*Card{},
	}

	for _, c := range list {
		p.Cards[c.OID] = c.clone()
	}

	return &p
}

func group(id string) types.Group {
	return types.Group{
		OID:  id,
		Name: "",
	}
}

func makeCard(oid catalog.OID, name string, card uint32, groups ...string) Card {
	n := types.Name(name)
	var c *types.Card

	if card > 0 {
		cc := types.Card(card)
		c = &cc
	}

	cardholder := Card{
		OID:    oid,
		Name:   &n,
		Card:   c,
		From:   date("2021-01-02"),
		To:     date("2021-12-30"),
		Groups: map[catalog.OID]bool{},
	}

	for _, g := range groups {
		cardholder.Groups[catalog.OID(g)] = true
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

func compareDB(db, expected *Cards, t *testing.T) {
	compare(db.Cards, expected.Cards, t)
}
