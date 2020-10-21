package memdb

import (
	"reflect"
	"testing"
	"time"

	"github.com/uhppoted/uhppoted-httpd/types"
)

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

func cardholder(id, name string, card uint32) *types.CardHolder {
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
			if !reflect.DeepEqual(*v, *w) {
				t.Errorf("Added cardholder %v does not match\n   expected:%v\n   got:     %v", w.ID, *w, *v)
			}
		}
	}

	if len(p.Updated) != len(q.Updated) {
		t.Errorf("Incorrect return 'updated' list\n   expected:%#v\n   got:     %#v", q, p)
	} else {
		for i := range q.Updated {
			v := p.Updated[i].(*types.CardHolder)
			w := q.Updated[i].(*types.CardHolder)

			if *v.Name != *w.Name {
				t.Errorf("Updated cardholder %v - name does not match\n   expected:%v\n   got:     %v", w.ID, *w.Name, *v.Name)
			}

			if *v.Card != *w.Card {
				t.Errorf("Updated cardholder %v - card does not match\n   expected:%v\n   got:     %v", w.ID, *w.Card, *v.Card)
			}

			if *v.From != *w.From {
				t.Errorf("Updated cardholder %v - 'from' date does not match\n   expected:%v\n   got:     %v", w.ID, *w.From, *v.From)
			}

			if *v.To != *w.To {
				t.Errorf("Updated cardholder %v - 'to' date does not match\n   expected:%v\n   got:     %v", w.ID, *w.To, *v.To)
			}

			if !reflect.DeepEqual(v.Groups, w.Groups) {
				t.Errorf("Updated cardholder %v - groups do not match\n   expected:%v\n   got:     %v", w.ID, w.Groups, v.Groups)
			}
		}
	}
}

func compareDB(db, expected *fdb, t *testing.T) {
	if !reflect.DeepEqual(db.data.Tables.Groups, expected.data.Tables.Groups) {
		t.Errorf("DB groups do not match\n   expected:%v\n   got:     %v", expected.data.Tables.Groups, db.data.Tables.Groups)
	}

	p := db.data.Tables.CardHolders
	q := db.data.Tables.CardHolders

	if len(p) != len(q) {
		t.Errorf("DB cardholders do not match\n   expected:%v\n   got:     %v", q, p)
	} else {
		for k, v := range q {
			w := p[k]
			if !reflect.DeepEqual(*v, *w) {
				t.Errorf("DB cardholder %v does not match\n   expected:%v\n   got:     %v", k, *w, *v)
			}
		}
	}
}
