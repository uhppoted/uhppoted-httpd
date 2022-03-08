package schema

import (
	"encoding/json"
	"testing"
)

func TestOIDAppend(t *testing.T) {
	tests := []struct {
		oid      OID
		suffix   Suffix
		expected OID
	}{
		{oid: OID("0.1"), suffix: "2.3", expected: OID("0.1.2.3")},
		{oid: OID("0.1."), suffix: "2.3", expected: OID("0.1.2.3")},
		{oid: OID("0.1"), suffix: ".2.3", expected: OID("0.1.2.3")},
		{oid: OID("0.1."), suffix: ".2.3", expected: OID("0.1.2.3")},
	}

	for _, v := range tests {
		joined := v.oid.Append(v.suffix)
		if joined != v.expected {
			t.Errorf("Incorrectly appended OID - expected:%v, got:%v", v.expected, joined)
		}
	}

}

func TestOIDAppendS(t *testing.T) {
	tests := []struct {
		oid      OID
		suffix   string
		expected OID
	}{
		{oid: OID("0.1"), suffix: "2.3", expected: OID("0.1.2.3")},
		{oid: OID("0.1."), suffix: "2.3", expected: OID("0.1.2.3")},
		{oid: OID("0.1"), suffix: ".2.3", expected: OID("0.1.2.3")},
		{oid: OID("0.1."), suffix: ".2.3", expected: OID("0.1.2.3")},
	}

	for _, v := range tests {
		joined := v.oid.AppendS(v.suffix)
		if joined != v.expected {
			t.Errorf("Incorrectly appended OID - expected:%v, got:%v", v.expected, joined)
		}
	}
}

func TestOIDMarshalJSON(t *testing.T) {
	oid := OID("0.1.2.3")
	b, err := json.Marshal(oid)
	if err != nil {
		t.Fatalf("Unexpected error marshaling OID (%v)", err)
	}

	if string(b) != `"0.1.2.3"` {
		t.Errorf("Invalid JSON OID - expected:%v, got:%v", "0.1.2.3", string(b))
	}
}

func TestOIDUnmarshalJSON(t *testing.T) {
	s := `"0.1.2.3"`
	expected := OID("0.1.2.3")

	var oid OID

	if err := json.Unmarshal([]byte(s), &oid); err != nil {
		t.Fatalf("Unexpected error unmarshaling OID (%v)", err)
	}

	if oid != expected {
		t.Errorf("Invalid OID - expected:%v, got:%v", expected, oid)
	}
}

func TestSuffixAppend(t *testing.T) {
	tests := []struct {
		suffix   Suffix
		s        string
		expected Suffix
	}{
		{suffix: "2.3", s: "4.5", expected: ".2.3.4.5"},
		{suffix: ".2.3", s: "4.5", expected: ".2.3.4.5"},
		{suffix: "2.3", s: ".4.5", expected: ".2.3.4.5"},
		{suffix: ".2.3", s: ".4.5", expected: ".2.3.4.5"},

		{suffix: "2.3.", s: "4.5", expected: ".2.3.4.5"},
		{suffix: ".2.3.", s: "4.5", expected: ".2.3.4.5"},
		{suffix: "2.3.", s: ".4.5", expected: ".2.3.4.5"},
		{suffix: ".2.3.", s: ".4.5", expected: ".2.3.4.5"},
	}

	for _, v := range tests {
		appended := v.suffix.Append(v.s)
		if appended != v.expected {
			t.Errorf("Incorrectly appended OID - expected:%v, got:%v", v.expected, appended)
		}
	}

}
