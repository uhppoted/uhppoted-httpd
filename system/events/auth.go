package events

import (
	"github.com/uhppoted/uhppoted-httpd/auth"
)

type TAuthable interface {
	Event

	AsRuleEntity() (string, any)
	CacheKey() string
}

var rulesets = []auth.RuleSet{auth.Events}

func CanView[T TAuthable](a auth.OpAuth, u T, field string, value any) error {
	return auth.CanView(a, u, field, value, rulesets...)
}

func CanAdd[T TAuthable](a auth.OpAuth, u T) error {
	return auth.CanAdd(a, u, rulesets...)
}

func CanUpdate[T TAuthable](a auth.OpAuth, u T, field string, value any) error {
	return auth.CanUpdate(a, u, field, value, rulesets...)
}

func CanDelete[T TAuthable](a auth.OpAuth, u T) error {
	return auth.CanDelete(a, u, rulesets...)
}
