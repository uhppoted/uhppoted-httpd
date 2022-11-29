package users

import (
	"github.com/uhppoted/uhppoted-httpd/auth"
)

type TAuthable interface {
	User

	AsRuleEntity() (string, any)
}

func CanView[T TAuthable](a auth.OpAuth, u T, field string, value any) error {
	if !auth.IsNil(a) {
		return a.CanView(u, field, value, auth.Users)
	}

	return nil
}
