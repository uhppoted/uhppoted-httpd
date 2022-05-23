package cards

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	core "github.com/uhppoted/uhppote-core/types"

	"github.com/uhppoted/uhppoted-httpd/audit"
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	"github.com/uhppoted/uhppoted-httpd/types"
)

var hagrid = makeCard("0.4.1", "Hagrid", 6514231)
var dobby = makeCard("0.4.2", "Dobby", 1234567, "G05")

type stub struct {
	canView       func(auth.RuleSet, auth.Operant, string, interface{}) error
	canUpdateCard func(auth.Operant, string, interface{}) error
	write         func(e audit.AuditRecord)
}

func (x *stub) CanView(object auth.Operant, field string, value interface{}, rulesets ...auth.RuleSet) error {
	if x.canView != nil && len(rulesets) > 0 {
		return x.canView(rulesets[0], object, field, value)
	}

	return nil
}

func (x *stub) CanAdd(operant auth.Operant, rulesets ...auth.RuleSet) error {
	return fmt.Errorf("not authorised")
}

func (x *stub) CanUpdate(operant auth.Operant, field string, value interface{}, rulesets ...auth.RuleSet) error {
	if len(rulesets) > 0 {
		switch rulesets[0] {
		case auth.Cards:
			if x.canUpdateCard != nil {
				return x.canUpdateCard(operant, field, value)
			}
		}
	}

	return fmt.Errorf("not authorised")
}

func (x *stub) CanDelete(operant auth.Operant, rulesets ...auth.RuleSet) error {
	return fmt.Errorf("not authorised")
}

func (x *stub) Write(e audit.AuditRecord) {
	x.write(e)
}

func date(s string) core.Date {
	date, _ := time.ParseInLocation("2006-01-02", s, time.Local)

	return core.Date(date)
}

func makeCards(list ...Card) *Cards {
	p := Cards{
		cards: map[schema.OID]*Card{},
	}

	for _, c := range list {
		p.cards[c.OID] = c.clone()
	}

	return &p
}

func group(id string) types.Group {
	return types.Group{
		OID:  id,
		Name: "",
	}
}

func makeCard(oid schema.OID, name string, card uint32, groups ...string) Card {
	cardholder := Card{
		CatalogCard: catalog.CatalogCard{
			OID:    oid,
			CardID: card,
		},
		name:   name,
		from:   date("2021-01-02"),
		to:     date("2021-12-30"),
		groups: map[schema.OID]bool{},
	}

	for _, g := range groups {
		cardholder.groups[schema.OID(g)] = true
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
	compare(db.cards, expected.cards, t)
}
