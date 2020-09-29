package memdb

import (
	"reflect"
	"testing"
	"time"

	"github.com/uhppoted/uhppoted-httpd/db"
	"github.com/uhppoted/uhppoted-httpd/types"
)

func date(s string) types.Date {
	date, _ := time.ParseInLocation("2006-01-02", s, time.Local)

	return types.Date(date)
}

func TestCardNumberUpdate(t *testing.T) {
	dbt := fdb{
		data: data{
			Tables: tables{
				Groups: []*db.Group{},
				CardHolders: []*db.CardHolder{
					&db.CardHolder{ID: "C01", Name: "Hagrid", Card: db.Card{ID: "CARD01", Number: 6514231}, From: date("2021-01-02"), To: date("2021-12-31"), Groups: []*db.Permission{}},
				},
			},
		},
	}

	u := map[string]interface{}{
		"CARD01": "1234567",
	}

	expected := struct {
		Updated map[string]interface{} `json:"updated"`
	}{
		Updated: map[string]interface{}{
			"CARD01": uint32(1234567),
		},
	}

	updated := fdb{
		data: data{
			Tables: tables{
				Groups: []*db.Group{},
				CardHolders: []*db.CardHolder{
					&db.CardHolder{ID: "C01", Name: "Hagrid", Card: db.Card{ID: "CARD01", Number: 1234567}, From: date("2021-01-02"), To: date("2021-12-31"), Groups: []*db.Permission{}},
				},
			},
		},
	}

	result, err := dbt.Update(u)
	if err != nil {
		t.Fatalf("Unexpected error updating DB: %v", err)
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Incorrect return value: expected:%#v, got:%#v", expected, result)
	}

	if !reflect.DeepEqual(dbt, updated) {
		t.Errorf("DB updated incorrectly: expected:%v, got:%v", updated, dbt)
	}
}

func TestDuplicateCardNumberUpdate(t *testing.T) {
	dbt := fdb{
		data: data{
			Tables: tables{
				Groups: []*db.Group{},
				CardHolders: []*db.CardHolder{
					&db.CardHolder{ID: "C01", Name: "Hagrid", Card: db.Card{ID: "CARD01", Number: 6514231}, From: date("2021-01-02"), To: date("2021-12-31"), Groups: []*db.Permission{}},
					&db.CardHolder{ID: "C02", Name: "Dobby", Card: db.Card{ID: "CARD02", Number: 1234567}, From: date("2021-01-02"), To: date("2021-12-31"), Groups: []*db.Permission{}},
				},
			},
		},
	}

	u := map[string]interface{}{
		"CARD01": "1234567",
	}

	updated := fdb{
		data: data{
			Tables: tables{
				Groups: []*db.Group{},
				CardHolders: []*db.CardHolder{
					&db.CardHolder{ID: "C01", Name: "Hagrid", Card: db.Card{ID: "CARD01", Number: 6514231}, From: date("2021-01-02"), To: date("2021-12-31"), Groups: []*db.Permission{}},
					&db.CardHolder{ID: "C02", Name: "Dobby", Card: db.Card{ID: "CARD02", Number: 1234567}, From: date("2021-01-02"), To: date("2021-12-31"), Groups: []*db.Permission{}},
				},
			},
		},
	}

	result, err := dbt.Update(u)
	if err == nil {
		t.Errorf("Expected error updating DB, got %v", err)
	}

	if result != nil {
		t.Errorf("Incorrect return value: expected:%#v, got:%#v", nil, result)
	}

	if !reflect.DeepEqual(dbt, updated) {
		t.Errorf("DB updated incorrectly: expected:%v, got:%v", updated, dbt)
	}
}

func TestCardNumberSwap(t *testing.T) {
	dbt := fdb{
		data: data{
			Tables: tables{
				Groups: []*db.Group{},
				CardHolders: []*db.CardHolder{
					&db.CardHolder{ID: "C01", Name: "Hagrid", Card: db.Card{ID: "CARD01", Number: 6514231}, From: date("2021-01-02"), To: date("2021-12-31"), Groups: []*db.Permission{}},
					&db.CardHolder{ID: "C02", Name: "Dobby", Card: db.Card{ID: "CARD02", Number: 1234567}, From: date("2021-01-02"), To: date("2021-12-31"), Groups: []*db.Permission{}},
				},
			},
		},
	}

	u := map[string]interface{}{
		"CARD01": "1234567",
		"CARD02": "6514231",
	}

	expected := struct {
		Updated map[string]interface{} `json:"updated"`
	}{
		Updated: map[string]interface{}{
			"CARD01": uint32(1234567),
			"CARD02": uint32(6514231),
		},
	}

	updated := fdb{
		data: data{
			Tables: tables{
				Groups: []*db.Group{},
				CardHolders: []*db.CardHolder{
					&db.CardHolder{ID: "C01", Name: "Hagrid", Card: db.Card{ID: "CARD01", Number: 1234567}, From: date("2021-01-02"), To: date("2021-12-31"), Groups: []*db.Permission{}},
					&db.CardHolder{ID: "C02", Name: "Dobby", Card: db.Card{ID: "CARD02", Number: 6514231}, From: date("2021-01-02"), To: date("2021-12-31"), Groups: []*db.Permission{}},
				},
			},
		},
	}

	result, err := dbt.Update(u)
	if err != nil {
		t.Fatalf("Unexpected error updating DB: %v", err)
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Incorrect return value: expected:%#v, got:%#v", expected, result)
	}

	if !reflect.DeepEqual(dbt, updated) {
		t.Errorf("DB updated incorrectly: expected:%v, got:%v", updated, dbt)
	}
}
