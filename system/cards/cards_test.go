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

	r, err := cards.UpdateByOID(nil, "<new>", "")
	if err != nil {
		t.Fatalf("Unexpected error adding new card (%v)", err)
	}

	if err := cards.Validate(); err != nil {
		t.Fatalf("Unexpected error validating cards with new card (%v)", err)
	}

	compare(r, expected, t)
	compareDB(cards, final, t)
}

func TestCardHolderAddWithAuth(t *testing.T) {
	cards := makeCards(hagrid)
	final := makeCards(hagrid)
	auth := stub{}

	catalog.PutCard(hagrid.OID)

	r, err := cards.UpdateByOID(&auth, "<new>", "")
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

// FIXME pending reworked implementation of 'add'
// func TestCardHolderAddWithInvalidGroup(t *testing.T) {
// 	dbt := dbx(hagrid)
// 	final := dbx(hagrid, cardholder("C02", "Dobby", 1234567))
//
// 	rq := map[string]interface{}{
// 		"cardholders": []map[string]interface{}{
// 			map[string]interface{}{
// 				"id":   "C02",
// 				"name": "Dobby",
// 				"card": 1234567,
// 				"from": "2021-01-02",
// 				"to":   "2021-12-30",
// 				"groups": map[string]bool{
// 					"G16": true,
// 				},
// 			},
// 		},
// 	}
//
// 	expected := result{
// 		Updated: []interface{}{
// 			cardholder("C02", "Dobby", 1234567),
// 		},
// 	}
//
// 	r, err := dbt.Post(rq, nil)
// 	if err != nil {
// 		t.Fatalf("Unexpected error adding card holder to DB: %v", err)
// 	}
//
// 	compare(r, expected, t)
// 	compareDB(dbt, final, t)
// }

// func TestCardHolderUpdate(t *testing.T) {
// 	dbt := dbx(hagrid)
// 	final := dbx(cardholder("C01", "Hagrid", 1234567))
//
// 	rq := map[string]interface{}{
// 		"cardholders": []map[string]interface{}{
// 			map[string]interface{}{
// 				"id":   "C01",
// 				"name": "Hagrid",
// 				"card": 1234567,
// 			},
// 		},
// 	}
//
// 	expected := result{
// 		Updated: []interface{}{
// 			cardholder("C01", "Hagrid", 1234567),
// 		},
// 	}
//
// 	r, err := dbt.Post(rq, nil)
// 	if err != nil {
// 		t.Fatalf("Unexpected error updating DB: %v", err)
// 	}
//
// 	compare(r, expected, t)
// 	compareDB(dbt, final, t)
// }

// func TestCardHolderUpdateWithAuth(t *testing.T) {
// 	dbt := dbx(hagrid)
// 	final := dbx(hagrid)
// 	auth := stub{}
//
// 	rq := map[string]interface{}{
// 		"cardholders": []map[string]interface{}{
// 			map[string]interface{}{
// 				"id":   "C01",
// 				"name": "Hagrid",
// 				"card": 1234567,
// 			},
// 		},
// 	}
//
// 	r, err := dbt.Post(rq, &auth)
//
// 	if err == nil {
// 		t.Errorf("Expected 'not authorised' error updating card holder in DB, got:%v", err)
// 	}
//
// 	if r != nil {
// 		t.Errorf("Unexpected return updating record without authorisation - expected:%v, got: %v", nil, err)
// 	}
//
// 	compareDB(dbt, final, t)
// }

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

	_, err := cards.UpdateByOID(nil, "0.3.1.2", "1234567")
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

	if _, err := cards.UpdateByOID(nil, "0.3.1.2", "1234567"); err != nil {
		t.Fatalf("Unexpected error updating cards (%v)", err)
	}

	if _, err := cards.UpdateByOID(nil, "0.3.2.2", "6514231"); err != nil {
		t.Fatalf("Unexpected error updating cards (%v)", err)
	}

	if err := cards.Validate(); err != nil {
		t.Fatalf("Unexpected error updating cards (%v)", err)
	}

	compareDB(cards, final, t)
}

func TestCardDelete(t *testing.T) {
	cards := makeCards(hagrid, dobby)

	catalog.PutCard(hagrid.OID)

	if _, err := cards.UpdateByOID(nil, dobby.OID.Append(catalog.CardName), ""); err != nil {
		t.Fatalf("Unexpected error deleting card (%v)", err)
	}

	if _, err := cards.UpdateByOID(nil, dobby.OID.Append(catalog.CardNumber), ""); err != nil {
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

	if _, err := cards.UpdateByOID(&authx, dobby.OID.Append(catalog.CardName), ""); err != nil {
		t.Fatalf("Unexpected error deleting card (%v)", err)
	}

	if _, err := cards.UpdateByOID(&authx, dobby.OID.Append(catalog.CardNumber), ""); err == nil {
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

// func TestCardHolderDeleteWithInvalidID(t *testing.T) {
// 	dbt := dbx(hagrid, dobby)
// 	final := dbx(hagrid, dobby)
//
// 	rq := map[string]interface{}{
// 		"cardholders": []map[string]interface{}{
// 			map[string]interface{}{
// 				"id":   "CXX",
// 				"name": "",
// 				"card": 0,
// 			},
// 		},
// 	}
//
// 	r, err := dbt.Post(rq, nil)
// 	if err == nil {
// 		t.Errorf("Expected error deleting non-existent record from DB - got: %v", err)
// 	}
//
// 	if r != nil {
// 		t.Errorf("Unexpected return deleting non-existent record from DB - expected:%v, got: %v", nil, err)
// 	}
//
// 	compareDB(dbt, final, t)
// }
