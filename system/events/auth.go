package events

import (
	"github.com/uhppoted/uhppoted-httpd/auth"
)

type TAuthable interface {
	Event

	AsRuleEntity() (string, any)
}

func CanView[T TAuthable](a auth.OpAuth, u T, field string, value any) error {
	if !auth.IsNil(a) {
		return a.CanView(u, field, value, auth.Events)
	}

	return nil
}

func CanUpdate[T TAuthable](a auth.OpAuth, u T, field string, value any) error {
	if !auth.IsNil(a) {
		return a.CanUpdate(u, field, value, auth.Events)
	}

	return nil
}

func CanDelete[T TAuthable](a auth.OpAuth, u T) error {
	if !auth.IsNil(a) {
		return a.CanDelete(u, auth.Events)
	}

	return nil
}
