package types

import (
	"testing"
)

func TestStatusToString(t *testing.T) {
	if s := StatusUnknown.String(); s != "unknown" {
		t.Errorf("Invalid string for StatusUnknown - expected:%s, got:%s", "unknown", s)
	}

	if s := StatusOk.String(); s != "ok" {
		t.Errorf("Invalid string for StatusOk - expected:%s, got:%s", "ok", s)
	}

	if s := StatusUncertain.String(); s != "uncertain" {
		t.Errorf("Invalid string for StatusUncertain - expected:%s, got:%s", "uncertain", s)
	}

	if s := StatusError.String(); s != "error" {
		t.Errorf("Invalid string for StatusError - expected:%s, got:%s", "error", s)
	}

	if s := StatusUnconfigured.String(); s != "unconfigured" {
		t.Errorf("Invalid string for StatusUnconfigured- expected:%s, got:%s", "unconfigured", s)
	}

	if s := StatusNew.String(); s != "new" {
		t.Errorf("Invalid string for StatusNew- expected:%s, got:%s", "new", s)
	}
}
