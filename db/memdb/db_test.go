package memdb

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/uhppoted/uhppoted-httpd/audit"
	"github.com/uhppoted/uhppoted-httpd/types"
)

var hagrid = cardholder("C01", "Hagrid", 6514231)
var dobby = cardholder("C02", "Dobby", 1234567, "G05")

func TestCardAdd(t *testing.T) {
	dbt := dbx(hagrid)
	final := dbx(hagrid, dobby)

	rq := map[string]interface{}{
		"cardholders": []map[string]interface{}{
			map[string]interface{}{
				"id":   "C02",
				"name": "Dobby",
				"card": 1234567,
				"from": "2021-01-02",
				"to":   "2021-12-30",
				"groups": map[string]bool{
					"G05": true,
				},
			},
		},
	}

	expected := result{
		Updated: []interface{}{
			cardholder("C02", "Dobby", 1234567, "G05"),
		},
	}

	r, err := dbt.Post(rq, nil)

	if err != nil {
		t.Fatalf("Unexpected error adding card holder to DB: %v", err)
	}

	compare(r, expected, t)
	compareDB(dbt, final, t)
}

func TestCardAddWithBlankNameAndCard(t *testing.T) {
	dbt := dbx(hagrid)
	final := dbx(cardholder("C01", "Hagrid", 6514231))

	rq := map[string]interface{}{
		"cardholders": []map[string]interface{}{
			map[string]interface{}{
				"id":   "C02",
				"from": "2021-01-02",
				"to":   "2021-12-30",
				"groups": map[string]bool{
					"G05": true,
				},
			},
		},
	}

	r, err := dbt.Post(rq, nil)

	if err == nil {
		t.Errorf("Expected error adding invalid card holder to DB, got:%v", err)
	}

	if r != nil {
		t.Errorf("Expected <nil> result adding invalid card holder to DB, got:%v", r)
	}

	compareDB(dbt, final, t)
}

func TestCardAddWithInvalidGroup(t *testing.T) {
	dbt := dbx(hagrid)
	final := dbx(hagrid, cardholder("C02", "Dobby", 1234567))

	rq := map[string]interface{}{
		"cardholders": []map[string]interface{}{
			map[string]interface{}{
				"id":   "C02",
				"name": "Dobby",
				"card": 1234567,
				"from": "2021-01-02",
				"to":   "2021-12-30",
				"groups": map[string]bool{
					"G16": true,
				},
			},
		},
	}

	expected := result{
		Updated: []interface{}{
			cardholder("C02", "Dobby", 1234567),
		},
	}

	r, err := dbt.Post(rq, nil)
	if err != nil {
		t.Fatalf("Unexpected error adding card holder to DB: %v", err)
	}

	compare(r, expected, t)
	compareDB(dbt, final, t)
}

func TestCardNumberUpdate(t *testing.T) {
	dbt := dbx(hagrid)
	final := dbx(cardholder("C01", "Hagrid", 1234567))

	rq := map[string]interface{}{
		"cardholders": []map[string]interface{}{
			map[string]interface{}{
				"id":   "C01",
				"name": "Hagrid",
				"card": 1234567,
			},
		},
	}

	expected := result{
		Updated: []interface{}{
			cardholder("C01", "Hagrid", 1234567),
		},
	}

	r, err := dbt.Post(rq, nil)
	if err != nil {
		t.Fatalf("Unexpected error updating DB: %v", err)
	}

	compare(r, expected, t)
	compareDB(dbt, final, t)
}

func TestDuplicateCardNumberUpdate(t *testing.T) {
	dbt := dbx(hagrid, dobby)
	final := dbx(hagrid, dobby)

	rq := map[string]interface{}{
		"cardholders": []map[string]interface{}{
			map[string]interface{}{
				"id":   "C01",
				"card": 1234567,
			},
		},
	}

	r, err := dbt.Post(rq, nil)
	if err == nil {
		t.Errorf("Expected error updating DB, got %v", err)
	}

	if r != nil {
		t.Errorf("Incorrect return value: expected:%#v, got:%#v", nil, r)
	}

	compareDB(dbt, final, t)
}

func TestCardNumberSwap(t *testing.T) {
	dbt := dbx(hagrid, dobby)
	final := dbx(cardholder("C01", "Hagrid", 1234567), cardholder("C02", "Dobby", 6514231, "G05"))

	rq := map[string]interface{}{
		"cardholders": []map[string]interface{}{
			map[string]interface{}{
				"id":   "C01",
				"name": "Hagrid",
				"card": 1234567,
			},
			map[string]interface{}{
				"id":   "C02",
				"name": "Dobby",
				"card": 6514231,
			},
		},
	}

	expected := result{
		Updated: []interface{}{
			cardholder("C01", "Hagrid", 1234567),
			cardholder("C02", "Dobby", 6514231, "G05"),
		},
	}

	r, err := dbt.Post(rq, nil)
	if err != nil {
		t.Fatalf("Unexpected error updating DB: %v", err)
	}

	compare(r, expected, t)
	compareDB(dbt, final, t)
}

func date(s string) *types.Date {
	date, _ := time.ParseInLocation("2006-01-02", s, time.Local)
	d := types.Date(date)

	return &d
}

func dbx(cardholders ...types.CardHolder) *fdb {
	p := fdb{
		data: data{
			Tables: tables{
				Groups: types.Groups{
					"G05": group("G05"),
				},
				CardHolders: types.CardHolders{},
			},
		},
		audit: audit.NewAuditTrail(),
	}

	for i, _ := range cardholders {
		c := cardholders[i].Clone()
		p.data.Tables.CardHolders[c.ID] = c
	}

	return &p
}

func group(id string) types.Group {
	return types.Group{
		ID:    id,
		Name:  "",
		Doors: []string{},
	}
}

func cardholder(id, name string, card uint32, groups ...string) types.CardHolder {
	n := types.Name(name)
	c := types.Card(card)

	cardholder := types.CardHolder{
		ID:     id,
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
