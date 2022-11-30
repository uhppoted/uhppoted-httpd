package users

import (
	"github.com/uhppoted/uhppoted-httpd/auth"
)

type TAuthable interface {
	User | *User

	AsRuleEntity() (string, any)
}

func CanView[T TAuthable](a auth.OpAuth, u T, field string, value any) error {
	if !auth.IsNil(a) {
		return a.CanView(u, field, value, auth.Users)
	}

	return nil
}

func CanUpdate[T TAuthable](a auth.OpAuth, u T, field string, value any) error {
	if !auth.IsNil(a) {
		return a.CanUpdate(u, field, value, auth.Users)
	}

	return nil
}

func CanDelete[T TAuthable](a auth.OpAuth, u T) error {
	if !auth.IsNil(a) {
		return a.CanDelete(u, auth.Users)
	}

	return nil
}
