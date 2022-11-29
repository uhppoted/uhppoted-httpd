package cards

import (
	"github.com/uhppoted/uhppoted-httpd/auth"
)

type TAuthable interface {
	Card

	AsRuleEntity() (string, any)
}

func CanView[T TAuthable](a auth.OpAuth, u T, field string, value any) error {
	if !auth.IsNil(a) {
		return a.CanView(u, field, value, auth.Cards)
	}

	return nil
}
