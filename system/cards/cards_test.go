package cards

import (
	"testing"

	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
)

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

// func TestAddCardWithAuditTrail(t *testing.T) {
//
// 	var logentry []byte
//
// 	expected := `{"UID":"","Module":"memdb","Operation":"add","Info":{"OID":"C02","Name":"Dobby","Card":1234567,"From":"2021-01-02","To":"2021-12-30","Groups":{"G05":true}}}`
// 	cards := makeCards(hagrid)
//
// 	trail = &stub{
// 		write: func(e audit.LogEntry) {
// 			logentry, _ = json.Marshal(e)
// 		},
// 	}
//
// 	catalog.PutCard(hagrid.OID)
//
// 	_, err := cards.UpdateByOID(nil, "<new>", "")
// 	if err != nil {
// 		t.Fatalf("Unexpected error adding new card (%v)", err)
// 	}
//
// 	if logentry == nil {
// 		t.Fatalf("Missing audit trail entry")
// 	}
//
// 	if string(logentry) != expected {
// 		t.Errorf("Incorrect audit trail record\n  expected:%+v\n  got:     %+v", expected, string(logentry))
// 	}
// }

func TestCardUpdate(t *testing.T) {
	cards := makeCards(hagrid)
	final := makeCards(makeCard(hagrid.OID, "Hagrid", 1234567))

	expected := []catalog.Object{
		catalog.Object{OID: "0.3.1.2", Value: "1234567"},
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

//func TestCardHolderUpdateWithAuditTrail(t *testing.T) {
//	var logentry []byte
//
//	dbt := dbx(hagrid)
//	trail = &stub{
//		write: func(e audit.LogEntry) {
//			logentry, _ = json.Marshal(e)
//		},
//	}
//
//	rq := map[string]interface{}{
//		"cardholders": []map[string]interface{}{
//			map[string]interface{}{
//				"id":   "C01",
//				"name": "Hagrid",
//				"card": 1234567,
//			},
//		},
//	}
//
//	dbt.Post(rq, nil)
//
//	expected := `{"UID":"","Module":"memdb","Operation":"update","Info":` +
//		`{"original":{"OID":"C01","Name":"Hagrid","Card":6514231,"From":"2021-01-02","To":"2021-12-30","Groups":{}},` +
//		`"updated":{"OID":"C01","Name":"Hagrid","Card":1234567,"From":"2021-01-02","To":"2021-12-30","Groups":{}}}}`
//
//	if logentry == nil {
//		t.Fatalf("Missing audit trail entry")
//	}
//
//	if string(logentry) != expected {
//		t.Errorf("Incorrect audit trail record\n  expected:%+v\n  got:     %+v", expected, string(logentry))
//	}
//}

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
		catalog.Object{OID: "0.3.1.5.10", Value: "true"},
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
		catalog.Object{OID: "0.3.1.5.10", Value: "false"},
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

// func TestCardHolderDeleteWithAuditTrail(t *testing.T) {
// 	var logentry []byte
//
// 	dbt := dbx(hagrid)
// 	trail = &stub{
// 		write: func(e audit.LogEntry) {
// 			logentry, _ = json.Marshal(e)
// 		},
// 	}
//
// 	rq := map[string]interface{}{
// 		"cardholders": []map[string]interface{}{
// 			map[string]interface{}{
// 				"id":   "C01",
// 				"name": "",
// 				"card": 0,
// 			},
// 		},
// 	}
//
// 	dbt.Post(rq, nil)
//
// 	expected := `{"UID":"","Module":"memdb","Operation":"delete","Info":{"OID":"C01","Name":"Hagrid","Card":6514231,"From":"2021-01-02","To":"2021-12-30","Groups":{}}}`
//
// 	if logentry == nil {
// 		t.Fatalf("Missing audit trail entry")
// 	}
//
// 	if string(logentry) != expected {
// 		t.Errorf("Incorrect audit trail record\n  expected:%+v\n  got:     %+v", expected, string(logentry))
// 	}
// }
