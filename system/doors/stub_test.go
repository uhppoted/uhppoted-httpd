package doors

import (
	"github.com/uhppoted/uhppoted-httpd/auth"
)

type stub struct {
	canView func(auth.RuleSet, auth.Operant, string, interface{}) error
}

func (x *stub) UID() string {
	return "stub"
}

func (x *stub) CanView(object auth.Operant, field string, value interface{}, rulesets ...auth.RuleSet) error {
	if x.canView != nil && len(rulesets) > 0 {
		return x.canView(rulesets[0], object, field, value)
	}

	return nil
}

func (x *stub) CanAdd(operant auth.Operant, rulesets ...auth.RuleSet) error {
	return nil
}

func (x *stub) CanUpdate(operant auth.Operant, field string, value interface{}, rulesets ...auth.RuleSet) error {
	return nil
}

func (x *stub) CanDelete(operant auth.Operant, rulesets ...auth.RuleSet) error {
	return nil
}
