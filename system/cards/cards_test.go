package cards

import (
	"reflect"
	"testing"

	"github.com/uhppoted/uhppoted-httpd/audit"
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
)

type dbc struct {
	logs []audit.AuditRecord
}

func (d *dbc) Write(record audit.AuditRecord) {
	d.logs = append(d.logs, record)
}

func (d *dbc) Stash(objects []catalog.Object) {
}

func (d *dbc) Objects() []catalog.Object {
	return []catalog.Object{}
}

func (d *dbc) Commit() {
}

func TestCardAdd(t *testing.T) {
	placeholder := Card{
		OID:    catalog.OID("0.3.2"),
		Groups: map[catalog.OID]bool{},
	}

	expected := []catalog.Object{
		catalog.Object{OID: "0.3.2", Value: "new"},
	}

	cards := makeCards(hagrid)
	final := makeCards(hagrid, placeholder)

	catalog.Clear()
	catalog.PutCard(hagrid.OID)

	r, err := cards.UpdateByOID(nil, "<new>", "", nil)
	if err != nil {
		t.Fatalf("Unexpected error adding new card (%v)", err)
	}

	if err := cards.Validate(); err != nil {
		t.Fatalf("Unexpected error validating cards with new card (%v)", err)
	}

	compare(r, expected, t)
	compareDB(cards, final, t)
}

func TestCardAddWithAuth(t *testing.T) {
	cards := makeCards(hagrid)
	final := makeCards(hagrid)
	auth := stub{}

	catalog.Clear()
	catalog.PutCard(hagrid.OID)

	r, err := cards.UpdateByOID(&auth, "<new>", "", nil)
	if err == nil {
		t.Errorf("Expected 'not authorised' error adding card, got:%v", err)
	}

	if r != nil {
		t.Errorf("Unexpected return adding card record without authorisation - expected:%v, got: %v", nil, err)
	}

	compareDB(cards, final, t)
}

func TestCardAddWithAuditTrail(t *testing.T) {
	trail := dbc{}

	expected := struct {
		returned []catalog.Object
		logs     []audit.AuditRecord
		db       *Cards
	}{
		returned: []catalog.Object{catalog.Object{OID: "0.3.2", Value: "new"}},

		logs: []audit.AuditRecord{
			audit.AuditRecord{
				UID:       "",
				OID:       "0.3.2",
				Component: "card",
				Operation: "add",
				Details: audit.Details{
					ID:          "",
					Name:        "",
					Field:       "card",
					Description: "Added <new> card",
					Before:      "",
					After:       "",
				},
			},
		},

		db: makeCards(hagrid, Card{
			OID:    catalog.OID("0.3.2"),
			Groups: map[catalog.OID]bool{},
		}),
	}

	cards := makeCards(hagrid)

	catalog.Clear()
	catalog.PutCard(hagrid.OID)

	r, err := cards.UpdateByOID(nil, "<new>", "", &trail)
	if err != nil {
		t.Fatalf("Unexpected error adding new card (%v)", err)
	}

	if err := cards.Validate(); err != nil {
		t.Fatalf("Unexpected error validating cards with new card (%v)", err)
	}

	compare(r, expected.returned, t)
	compareDB(cards, expected.db, t)

	if trail.logs == nil || len(trail.logs) != 1 {
		t.Error("Invalid audit trail")
	} else if !reflect.DeepEqual(trail.logs, expected.logs) {
		t.Errorf("Incorrect audit trail record\n  expected:%+v\n  got:     %+v", expected.logs, trail.logs)
	}
}

func TestCardUpdate(t *testing.T) {
	cards := makeCards(hagrid)
	final := makeCards(makeCard(hagrid.OID, "Hagrid", 1234567))

	expected := []catalog.Object{
		catalog.Object{OID: "0.3.1.2", Value: 1234567},
	}

	objects, err := cards.UpdateByOID(nil, hagrid.OID.Append(CardNumber), "1234567", nil)
	if err != nil {
		t.Errorf("Unexpected error updating card (%v)", err)
	}

	if err := cards.Validate(); err != nil {
		t.Errorf("Expected error updating card, got %v", err)
	}

	compare(objects, expected, t)
	compareDB(cards, final, t)
}

func TestCardUpdateWithInvalidOID(t *testing.T) {
	cards := makeCards(hagrid)
	final := makeCards(hagrid)
	expected := []catalog.Object{}

	objects, err := cards.UpdateByOID(nil, "0.3.5.2", "1234567", nil)
	if err != nil {
		t.Errorf("Unexpected error updating card (%v)", err)
	}

	if err := cards.Validate(); err != nil {
		t.Errorf("Expected error updating card, got %v", err)
	}

	compare(objects, expected, t)
	compareDB(cards, final, t)
}

func TestCardUpdateWithAuth(t *testing.T) {
	cards := makeCards(hagrid)
	final := makeCards(hagrid)
	auth := stub{}

	if _, err := cards.UpdateByOID(&auth, hagrid.OID.Append(CardNumber), "1234567", nil); err == nil {
		t.Errorf("Expected 'not authorised' error updating card, got:%v", err)
	}

	compareDB(cards, final, t)
}

func TestCardUpdateWithAuditTrail(t *testing.T) {
	trail := dbc{}
	cards := makeCards(hagrid)
	expected := struct {
		returned []catalog.Object
		logs     []audit.AuditRecord
		db       *Cards
	}{
		returned: []catalog.Object{catalog.Object{OID: "0.3.1.2", Value: 1234567}},

		logs: []audit.AuditRecord{
			audit.AuditRecord{
				UID:       "",
				OID:       "0.3.1",
				Component: "card",
				Operation: "update",
				Details: audit.Details{
					ID:          "6514231",
					Name:        "Hagrid",
					Field:       "card",
					Description: "Updated card number from 6514231 to 1234567",
					Before:      "6514231",
					After:       "1234567",
				},
			},
		},

		db: makeCards(makeCard(hagrid.OID, "Hagrid", 1234567)),
	}

	objects, err := cards.UpdateByOID(nil, hagrid.OID.Append(CardNumber), "1234567", &trail)
	if err != nil {
		t.Errorf("Unexpected error updating card (%v)", err)
	}

	if err := cards.Validate(); err != nil {
		t.Errorf("Expected error updating card, got %v", err)
	}

	compare(objects, expected.returned, t)
	compareDB(cards, expected.db, t)

	if trail.logs == nil || len(trail.logs) != 1 {
		t.Error("Invalid audit trail")
	} else if !reflect.DeepEqual(trail.logs, expected.logs) {
		t.Errorf("Incorrect audit trail record\n  expected:%+v\n  got:     %+v", expected.logs, trail.logs)
	}
}

func TestDuplicateCardNumberUpdate(t *testing.T) {
	cards := makeCards(hagrid, dobby)

	_, err := cards.UpdateByOID(nil, "0.3.1.2", "1234567", nil)
	if err != nil {
		t.Errorf("Unexpected error updating cards (%v)", err)
	}

	if err := cards.Validate(); err == nil {
		t.Errorf("Expected error updating cards, got %v", err)
	}
}

func TestCardNumberSwap(t *testing.T) {
	cards := makeCards(hagrid, dobby)
	final := makeCards(makeCard("0.3.1", "Hagrid", 1234567), makeCard("0.3.2", "Dobby", 6514231, "G05"))

	if _, err := cards.UpdateByOID(nil, "0.3.1.2", "1234567", nil); err != nil {
		t.Fatalf("Unexpected error updating cards (%v)", err)
	}

	if _, err := cards.UpdateByOID(nil, "0.3.2.2", "6514231", nil); err != nil {
		t.Fatalf("Unexpected error updating cards (%v)", err)
	}

	if err := cards.Validate(); err != nil {
		t.Fatalf("Unexpected error updating cards (%v)", err)
	}

	compareDB(cards, final, t)
}

func TestCardUpdateAddGroup(t *testing.T) {
	catalog.PutGroup(catalog.OID("0.4.10"))

	cards := makeCards(hagrid)
	final := makeCards(makeCard(hagrid.OID, "Hagrid", 6514231, "0.4.10"))
	expected := []catalog.Object{
		catalog.Object{OID: "0.3.1.5.10", Value: true},
	}

	objects, err := cards.UpdateByOID(nil, catalog.OID("0.3.1.5.10"), "true", nil)
	if err != nil {
		t.Errorf("Unexpected error updating card [%v]", err)
	}

	if err := cards.Validate(); err != nil {
		t.Errorf("Expected error updating card, got %v", err)
	}

	compare(objects, expected, t)
	compareDB(cards, final, t)
}

func TestCardUpdateRemoveGroup(t *testing.T) {
	catalog.PutGroup(catalog.OID("0.4.10"))

	hagrid2 := makeCard(hagrid.OID, "Hagrid", 6514231)
	hagrid2.Groups["0.4.10"] = false
	cards := makeCards(hagrid)
	final := makeCards(hagrid2)
	expected := []catalog.Object{
		catalog.Object{OID: "0.3.1.5.10", Value: false},
	}

	objects, err := cards.UpdateByOID(nil, catalog.OID("0.3.1.5.10"), "false", nil)
	if err != nil {
		t.Errorf("Unexpected error updating card (%v)", err)
	}

	if err := cards.Validate(); err != nil {
		t.Errorf("Expected error updating card, got %v", err)
	}

	compare(objects, expected, t)
	compareDB(cards, final, t)
}

func TestCardUpdateWithInvalidGroup(t *testing.T) {
	cards := makeCards(hagrid)
	final := makeCards(hagrid)

	objects, err := cards.UpdateByOID(nil, catalog.OID("0.3.1.5.99"), "true", nil)
	if err == nil {
		t.Errorf("Expected error updating card, got:%v", err)
	}

	compare(objects, nil, t)
	compareDB(cards, final, t)
}

func TestCardDelete(t *testing.T) {
	cards := makeCards(hagrid, dobby)

	catalog.PutCard(hagrid.OID)

	if _, err := cards.UpdateByOID(nil, dobby.OID.Append(catalog.CardName), "", nil); err != nil {
		t.Fatalf("Unexpected error deleting card (%v)", err)
	}

	if _, err := cards.UpdateByOID(nil, dobby.OID.Append(catalog.CardNumber), "", nil); err != nil {
		t.Fatalf("Unexpected error deleting card (%v)", err)
	}

	if err := cards.Validate(); err != nil {
		t.Fatalf("Unexpected error validating cards with deleted card (%v)", err)
	}

	if cards.Cards[dobby.OID].deleted == nil {
		t.Errorf("Failed to mark card %v as 'deleted'", dobby.Card)
	}
}

func TestCardHolderDeleteWithAuth(t *testing.T) {
	cards := makeCards(hagrid, dobby)
	authx := stub{
		canUpdateCard: func(card auth.Operant, field string, value interface{}) error {
			return nil
		},
	}

	if _, err := cards.UpdateByOID(&authx, dobby.OID.Append(catalog.CardName), "", nil); err != nil {
		t.Fatalf("Unexpected error deleting card (%v)", err)
	}

	if _, err := cards.UpdateByOID(&authx, dobby.OID.Append(catalog.CardNumber), "", nil); err == nil {
		t.Fatalf("Expected 'not authorised' error deleting card, got:%v", err)
	}
}

func TestCardHolderDeleteWithAuditTrail(t *testing.T) {
	trail := dbc{}

	expected := struct {
		logs []audit.AuditRecord
		db   *Cards
	}{
		logs: []audit.AuditRecord{
			audit.AuditRecord{
				UID:       "",
				OID:       "0.3.2",
				Component: "card",
				Operation: "update",
				Details: audit.Details{
					ID:          "1234567",
					Name:        "Dobby",
					Field:       "name",
					Description: "Updated name from Dobby to <blank>",
					Before:      "Dobby",
					After:       "",
				},
			},
			audit.AuditRecord{
				UID:       "",
				OID:       "0.3.2",
				Component: "card",
				Operation: "update",
				Details: audit.Details{
					ID:          "1234567",
					Name:        "",
					Field:       "number",
					Description: "Cleared card number 1234567",
					Before:      "1234567",
					After:       "",
				},
			},
			audit.AuditRecord{
				UID:       "",
				OID:       "0.3.2",
				Component: "card",
				Operation: "delete",
				Details: audit.Details{
					ID:          "",
					Name:        "",
					Field:       "card",
					Description: "Deleted card 1234567",
					Before:      "",
					After:       "",
				},
			},
		},

		db: makeCards(hagrid, makeCard(dobby.OID, "", 0, "G05")),
	}

	cards := makeCards(hagrid, dobby)

	catalog.Clear()
	catalog.PutCard(hagrid.OID)
	catalog.PutCard(dobby.OID)

	cards.UpdateByOID(nil, dobby.OID.Append(catalog.CardName), "", &trail)
	cards.UpdateByOID(nil, dobby.OID.Append(catalog.CardNumber), "", &trail)

	compareDB(cards, expected.db, t)

	if trail.logs == nil {
		t.Error("Invalid audit trail")
	} else if !reflect.DeepEqual(trail.logs, expected.logs) {
		t.Errorf("Incorrect audit trail record\n  expected:%+v\n  got:     %+v", expected.logs, trail.logs)
	}
}
