package cards

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/uhppoted/uhppoted-httpd/audit"
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/impl"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	"github.com/uhppoted/uhppoted-httpd/system/db"
	"github.com/uhppoted/uhppoted-httpd/system/groups"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type dbc struct {
	logs []audit.AuditRecord
}

func (d *dbc) Write(record audit.AuditRecord) {
	d.logs = append(d.logs, record)
}

func (d *dbc) Stash(objects []schema.Object) {
}

func (d *dbc) Updated(oid schema.OID, field schema.Suffix, value any) {
}

func (d *dbc) Objects() []schema.Object {
	return []schema.Object{}
}

func (d *dbc) Commit(sys db.System, hook func()) {
}

func (d *dbc) SetPassword(uid, pwd, role string) error {
	return nil
}

func TestCardAdd(t *testing.T) {
	catalog.Init(memdb.NewCatalog())

	expected := []schema.Object{
		schema.Object{OID: "0.4.2", Value: "new"},
		schema.Object{OID: "0.4.2.0.1", Value: types.TimestampNow()},
	}

	cards := makeCards(hagrid)
	final := makeCards(hagrid, Card{
		CatalogCard: catalog.CatalogCard{
			OID: "0.4.2",
		},
	})

	catalog.PutT(hagrid.CatalogCard)

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
	catalog.Init(memdb.NewCatalog())

	cards := makeCards(hagrid)
	final := makeCards(hagrid)
	a := auth.Authorizator{
		OpAuth: &stub{},
	}

	catalog.PutT(hagrid.CatalogCard)

	r, err := cards.UpdateByOID(&a, "<new>", "", nil)
	if err == nil {
		t.Errorf("Expected 'not authorised' error adding card, got:%v", err)
	}

	if r != nil {
		t.Errorf("Unexpected return adding card record without authorisation - expected:%v, got: %v", nil, err)
	}

	compareDB(cards, final, t)
}

func TestCardAddWithAuditTrail(t *testing.T) {
	catalog.Init(memdb.NewCatalog())

	trail := dbc{}

	expected := struct {
		returned []schema.Object
		logs     []audit.AuditRecord
		db       *Cards
	}{
		returned: []schema.Object{
			schema.Object{OID: "0.4.2", Value: "new"},
			schema.Object{OID: "0.4.2.0.1", Value: types.TimestampNow()},
		},

		logs: []audit.AuditRecord{
			audit.AuditRecord{
				UID:       "",
				OID:       "0.4.2",
				Component: "card",
				Operation: "add",
				Details: audit.Details{
					ID:          "",
					Name:        "",
					Field:       "card",
					Description: "Added 'new' card",
					Before:      "",
					After:       "",
				},
			},
		},

		db: makeCards(hagrid, Card{
			CatalogCard: catalog.CatalogCard{
				OID: "0.4.2",
			},
		}),
	}

	cards := makeCards(hagrid)

	catalog.PutT(hagrid.CatalogCard)

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

	expected := []schema.Object{
		{OID: "0.4.1", Value: ""},
		{OID: "0.4.1.2", Value: 1234567},
		{OID: "0.4.1.0.0", Value: types.StatusOk},
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
	expected := []schema.Object{}

	objects, err := cards.UpdateByOID(nil, "0.4.5.2", "1234567", nil)
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
	auth := auth.Authorizator{
		OpAuth: &stub{},
	}

	if _, err := cards.UpdateByOID(&auth, hagrid.OID.Append(CardNumber), "1234567", nil); err == nil {
		t.Errorf("Expected 'not authorised' error updating card, got:%v", err)
	}

	compareDB(cards, final, t)
}

func TestCardUpdateWithAuditTrail(t *testing.T) {
	trail := dbc{}
	cards := makeCards(hagrid)
	expected := struct {
		returned []schema.Object
		logs     []audit.AuditRecord
		db       *Cards
	}{
		returned: []schema.Object{
			{OID: "0.4.1", Value: ""},
			{OID: "0.4.1.2", Value: 1234567},
			{OID: "0.4.1.0.0", Value: types.StatusOk},
		},

		logs: []audit.AuditRecord{
			audit.AuditRecord{
				UID:       "",
				OID:       "0.4.1",
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

	_, err := cards.UpdateByOID(nil, "0.4.1.2", "1234567", nil)
	if err != nil {
		t.Errorf("Unexpected error updating cards (%v)", err)
	}

	if err := cards.Validate(); err == nil {
		t.Errorf("Expected error updating cards, got %v", err)
	}
}

func TestCardNumberSwap(t *testing.T) {
	cards := makeCards(hagrid, dobby)
	final := makeCards(makeCard("0.4.1", "Hagrid", 1234567), makeCard("0.4.2", "Dobby", 6514231, "G05"))

	if _, err := cards.UpdateByOID(nil, "0.4.1.2", "1234567", nil); err != nil {
		t.Fatalf("Unexpected error updating cards (%v)", err)
	}

	if _, err := cards.UpdateByOID(nil, "0.4.2.2", "6514231", nil); err != nil {
		t.Fatalf("Unexpected error updating cards (%v)", err)
	}

	if err := cards.Validate(); err != nil {
		t.Fatalf("Unexpected error updating cards (%v)", err)
	}

	compareDB(cards, final, t)
}

func TestCardUpdateAddGroup(t *testing.T) {
	oid := schema.GroupsOID.Append("10")
	group := groups.Group{
		CatalogGroup: catalog.CatalogGroup{
			OID: oid,
		},
	}

	catalog.PutT(group.CatalogGroup)

	cards := makeCards(hagrid)
	final := makeCards(makeCard(hagrid.OID, "Hagrid", 6514231, fmt.Sprintf("%v", oid)))
	expected := []schema.Object{
		{OID: "0.4.1", Value: ""},
		{OID: "0.4.1.5.10", Value: true},
		{OID: "0.4.1.0.0", Value: types.StatusOk},
	}

	objects, err := cards.UpdateByOID(nil, schema.OID("0.4.1.5.10"), "true", nil)
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
	oid := schema.GroupsOID.Append("10")
	group := groups.Group{
		CatalogGroup: catalog.CatalogGroup{
			OID: oid,
		},
	}
	catalog.PutT(group.CatalogGroup)

	hagrid2 := makeCard(hagrid.OID, "Hagrid", 6514231)
	hagrid2.groups[oid] = false
	cards := makeCards(hagrid)
	final := makeCards(hagrid2)
	expected := []schema.Object{
		{OID: "0.4.1", Value: ""},
		{OID: "0.4.1.5.10", Value: false},
		{OID: "0.4.1.0.0", Value: types.StatusOk},
	}

	objects, err := cards.UpdateByOID(nil, schema.OID("0.4.1.5.10"), "false", nil)
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

	objects, err := cards.UpdateByOID(nil, schema.OID("0.4.1.5.99"), "true", nil)
	if err == nil {
		t.Errorf("Expected error updating card, got:%v", err)
	}

	compare(objects, nil, t)
	compareDB(cards, final, t)
}

func TestCardDelete(t *testing.T) {
	cards := makeCards(hagrid, dobby)

	catalog.PutT(hagrid.CatalogCard)

	if _, err := cards.DeleteByOID(nil, dobby.OID, nil); err != nil {
		t.Fatalf("Unexpected error deleting card (%v)", err)
	}

	if err := cards.Validate(); err != nil {
		t.Fatalf("Unexpected error validating cards with deleted card (%v)", err)
	}

	if !cards.cards[dobby.OID].IsDeleted() {
		t.Errorf("Failed to mark card %v as 'deleted'", dobby.card)
	}
}

func TestCardHolderDeleteWithAuth(t *testing.T) {
	cards := makeCards(hagrid, dobby)
	auth := auth.Authorizator{
		OpAuth: &stub{
			canUpdateCard: func(card auth.Operant, field string, value interface{}) error {
				return nil
			},
		},
	}

	if _, err := cards.DeleteByOID(&auth, dobby.OID, nil); err == nil {
		t.Fatalf("Expected 'not authorised' error deleting card, got:%v", err)
	}
}

func TestCardHolderDeleteWithAuditTrail(t *testing.T) {
	catalog.Init(memdb.NewCatalog())

	trail := dbc{}

	expected := struct {
		logs []audit.AuditRecord
		db   *Cards
	}{
		logs: []audit.AuditRecord{
			audit.AuditRecord{
				UID:       "",
				OID:       "0.4.2",
				Component: "card",
				Operation: "delete",
				Details: audit.Details{
					ID:          "1234567",
					Name:        "Dobby",
					Field:       "card",
					Description: "Deleted card 1234567",
					Before:      "",
					After:       "",
				},
			},
		},

		db: makeCards(hagrid, dobby),
	}

	cards := makeCards(hagrid, dobby)

	catalog.PutT(hagrid.CatalogCard)
	catalog.PutT(dobby.CatalogCard)

	cards.DeleteByOID(nil, dobby.OID, &trail)

	compareDB(cards, expected.db, t)

	if trail.logs == nil {
		t.Error("Invalid audit trail")
	} else if !reflect.DeepEqual(trail.logs, expected.logs) {
		t.Errorf("Incorrect audit trail record\n  expected:%+v\n  got:     %+v", expected.logs, trail.logs)
	}
}

func TestValidateWithInvalidCard(t *testing.T) {
	cc := Cards{
		cards: map[schema.OID]*Card{
			"0.4.3": &Card{
				CatalogCard: catalog.CatalogCard{
					OID: "0.4.3",
				},
				name:     "",
				card:     0,
				created:  types.TimestampNow(),
				modified: types.TimestampNow(),
			},
		},
	}

	if err := cc.Validate(); err == nil {
		t.Errorf("Expected error validating cards list with invalid card (%v)", err)
	}
}

func TestValidateWithNewCard(t *testing.T) {
	cc := Cards{
		cards: map[schema.OID]*Card{
			"0.4.3": &Card{
				CatalogCard: catalog.CatalogCard{
					OID: "0.4.3",
				},
				name:    "",
				card:    0,
				created: types.TimestampNow(),
			},
		},
	}

	if err := cc.Validate(); err != nil {
		t.Errorf("Unexpected error validating cards list with new card (%v)", err)
	}
}
