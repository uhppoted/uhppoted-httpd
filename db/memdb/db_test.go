package memdb

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/uhppoted/uhppoted-httpd/types"
)

func TestCardAdd(t *testing.T) {
	dbt := fdb{
		data: data{
			Tables: tables{
				Groups: types.Groups{
					"G05": types.Group{},
				},
				CardHolders: types.CardHolders{
					"C01": cardholder("C01", "Hagrid", 6514231),
				},
			},
		},
	}

	final := fdb{
		data: data{
			Tables: tables{
				Groups: types.Groups{
					"G05": types.Group{},
				},
				CardHolders: types.CardHolders{
					"C01": cardholder("C01", "Hagrid", 6514231),
					"C02": cardholder("C02", "Dobby", 1234567, "G05"),
				},
			},
		},
	}

	expected := result{
		Added: []interface{}{
			cardholder("C02", "Dobby", 1234567, "G05"),
		},
	}

	r, err := dbt.add("C02", map[string]interface{}{
		"name": "Dobby",
		"card": 1234567,
		"from": "2021-01-02",
		"to":   "2021-12-30",
		"groups": map[string]bool{
			"G05": true,
		},
	})

	if err != nil {
		t.Fatalf("Unexpected error adding card holder to DB: %v", err)
	}

	compare(r, &expected, t)
	compareDB(&dbt, &final, t)
}

func TestCardAddWithBlankNameAndCard(t *testing.T) {
	dbt := fdb{
		data: data{
			Tables: tables{
				Groups: types.Groups{
					"G05": types.Group{},
				},
				CardHolders: types.CardHolders{
					"C01": cardholder("C01", "Hagrid", 6514231),
				},
			},
		},
	}

	final := fdb{
		data: data{
			Tables: tables{
				Groups: types.Groups{
					"G05": types.Group{},
				},
				CardHolders: types.CardHolders{
					"C01": cardholder("C01", "Hagrid", 6514231),
				},
			},
		},
	}

	r, err := dbt.add("C02", map[string]interface{}{
		"from": "2021-01-02",
		"to":   "2021-12-30",
		"groups": map[string]bool{
			"G05": true,
		},
	})

	if err == nil {
		t.Errorf("Expected error adding invalid card holder to DB, got:%v", err)
	}

	if r != nil {
		t.Errorf("Expected <nil> result adding invalid card holder to DB, got:%v", r)
	}

	compareDB(&dbt, &final, t)
}

func TestCardAddWithInvalidGroup(t *testing.T) {
	dbt := fdb{
		data: data{
			Tables: tables{
				Groups: types.Groups{
					"G05": types.Group{},
				},
				CardHolders: types.CardHolders{
					"C01": cardholder("C01", "Hagrid", 6514231),
				},
			},
		},
	}

	final := fdb{
		data: data{
			Tables: tables{
				Groups: types.Groups{
					"G05": types.Group{},
				},
				CardHolders: types.CardHolders{
					"C01": cardholder("C01", "Hagrid", 6514231),
					"C02": cardholder("C02", "Dobby", 1234567),
				},
			},
		},
	}

	expected := result{
		Added: []interface{}{
			cardholder("C02", "Dobby", 1234567),
		},
	}

	r, err := dbt.add("C02", map[string]interface{}{
		"name": "Dobby",
		"card": 1234567,
		"from": "2021-01-02",
		"to":   "2021-12-30",
		"groups": map[string]bool{
			"G16": true,
		},
	})
	if err != nil {
		t.Fatalf("Unexpected error adding card holder to DB: %v", err)
	}

	compare(r, &expected, t)
	compareDB(&dbt, &final, t)
}

func TestCardNumberUpdate(t *testing.T) {
	dbt := fdb{
		data: data{
			Tables: tables{
				Groups: types.Groups{},
				CardHolders: types.CardHolders{
					"C01": cardholder("C01", "Hagrid", 6514231),
				},
			},
		},
	}

	final := fdb{
		data: data{
			Tables: tables{
				Groups: types.Groups{},
				CardHolders: types.CardHolders{
					"C01": cardholder("C01", "Hagrid", 1234567),
				},
			},
		},
	}

	expected := result{
		Updated: []interface{}{
			cardholder("C01", "Hagrid", 1234567),
		},
	}

	r, err := dbt.update("C01", map[string]interface{}{"card": 1234567})
	if err != nil {
		t.Fatalf("Unexpected error updating DB: %v", err)
	}

	compare(r, &expected, t)
	compareDB(&dbt, &final, t)
}

func TestDuplicateCardNumberUpdate(t *testing.T) {
	dbt := fdb{
		data: data{
			Tables: tables{
				Groups: types.Groups{},
				CardHolders: types.CardHolders{
					"C01": cardholder("C01", "Hagrid", 6514231),
					"C02": cardholder("C02", "Dobby", 1234567),
				},
			},
		},
	}

	final := fdb{
		data: data{
			Tables: tables{
				Groups: types.Groups{},
				CardHolders: types.CardHolders{
					"C01": cardholder("C01", "Hagrid", 6514231),
					"C02": cardholder("C02", "Dobby", 1234567),
				},
			},
		},
	}

	r, err := dbt.update("C01", map[string]interface{}{"card": 1234567})
	if err == nil {
		t.Errorf("Expected error updating DB, got %v", err)
	}

	if r != nil {
		t.Errorf("Incorrect return value: expected:%#v, got:%#v", nil, r)
	}

	compareDB(&dbt, &final, t)
}

func TestCardNumberSwap(t *testing.T) {
	t.Skip() // FIXME DOESN'T WORK ANY MORE
	dbt := fdb{
		data: data{
			Tables: tables{
				Groups: types.Groups{},
				CardHolders: types.CardHolders{
					"C01": cardholder("C01", "Hagrid", 6514231),
					"C02": cardholder("C02", "Dobby", 1234567),
				},
			},
		},
	}

	final := fdb{
		data: data{
			Tables: tables{
				Groups: types.Groups{},
				CardHolders: types.CardHolders{
					"C01": cardholder("C01", "Hagrid", 1234567),
					"C02": cardholder("C02", "Dobby", 6514231),
				},
			},
		},
	}

	expected := result{
		Updated: []interface{}{
			cardholder("C01", "Hagrid", 1234567),
			cardholder("C02", "Dobby", 65414231),
		},
	}

	r, err := dbt.update("C01", map[string]interface{}{
		"C01": map[string]interface{}{"card": 1234567},
		"C02": map[string]interface{}{"card": 6514231},
	})
	if err != nil {
		t.Fatalf("Unexpected error updating DB: %v", err)
	}

	compare(r, &expected, t)
	compareDB(&dbt, &final, t)
}

func date(s string) *types.Date {
	date, _ := time.ParseInLocation("2006-01-02", s, time.Local)
	d := types.Date(date)

	return &d
}

func cardholder(id, name string, card uint32, groups ...string) *types.CardHolder {
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

	return &cardholder
}

func compare(got, expected interface{}, t *testing.T) {
	p := got.(result)
	q := expected.(*result)

	if len(p.Added) != len(q.Added) {
		t.Errorf("Incorrect return 'added' list\n   expected:%#v\n   got:     %#v", q, p)
	} else {
		for i := range q.Added {
			v := p.Added[i].(*types.CardHolder)
			w := q.Added[i].(*types.CardHolder)

			compareCardHolder(v, w, t)
		}
	}

	if len(p.Updated) != len(q.Updated) {
		t.Errorf("Incorrect return 'updated' list\n   expected:%#v\n   got:     %#v", q, p)
	} else {
		for i := range q.Updated {
			v := p.Updated[i].(*types.CardHolder)
			w := q.Updated[i].(*types.CardHolder)

			compareCardHolder(v, w, t)
		}
	}
}

func compareDB(db, expected *fdb, t *testing.T) {
	g := fmt.Sprint(db.data.Tables.Groups)
	h := fmt.Sprint(expected.data.Tables.Groups)
	if g != h {
		t.Errorf("DB groups do not match\n   expected:%v\n   got:     %v", h, g)
	}

	p := db.data.Tables.CardHolders
	q := expected.data.Tables.CardHolders

	if len(p) != len(q) {
		t.Errorf("DB cardholders do not match\n   expected:%v\n   got:     %v", q, p)
	} else {
		for k, w := range q {
			v := p[k]
			compareCardHolder(v, w, t)
		}
	}
}

func compareCardHolder(got, expected *types.CardHolder, t *testing.T) {
	if *got.Name != *expected.Name {
		t.Errorf("Updated cardholder %v - name does not match\n   expected:%v\n   got:     %v", expected.ID, *expected.Name, *got.Name)
	}

	if *got.Card != *expected.Card {
		t.Errorf("Updated cardholder %v - card does not match\n   expected:%v\n   got:     %v", expected.ID, *expected.Card, *got.Card)
	}

	if got.From != expected.From {
		if got.From != nil && expected.From != nil {
			if *got.From != *expected.From {
				t.Errorf("Updated cardholder %v - 'from' date does not match\n   expected:%v\n   got:     %v", expected.ID, expected.From.String(), got.From.String())
			}
		} else {
			t.Errorf("Updated cardholder %v - 'from' date does not match\n   expected:%v\n   got:     %v", expected.ID, expected.From, got.From)
		}
	}

	if got.To != expected.To {
		if got.To != nil && expected.To != nil {
			if *got.To != *expected.To {
				t.Errorf("Updated cardholder %v - 'to' date does not match\n   expected:%v\n   got:     %v", expected.ID, expected.To.String(), got.To.String())
			}
		} else {
			t.Errorf("Updated cardholder %v - 'to' date does not match\n   expected:%v\n   got:     %v", expected.ID, expected.To, got.To)
		}
	}

	if !reflect.DeepEqual(got.Groups, expected.Groups) {
		t.Errorf("Updated cardholder %v - groups do not match\n   expected:%v\n   got:     %v", expected.ID, expected.Groups, got.Groups)
	}
}
