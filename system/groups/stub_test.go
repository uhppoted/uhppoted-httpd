package groups

import (
	"github.com/uhppoted/uhppoted-httpd/auth"
)

type stub struct {
	canView func(auth.RuleSet, auth.Operant, string, any) error
}

func (x *stub) CanView(operant auth.Operant, field string, value any, rulesets ...auth.RuleSet) error {
	if x.canView != nil && len(rulesets) > 0 {
		return x.canView(rulesets[0], operant, field, value)
	}

	return nil
}

func (x *stub) CanAdd(operant auth.Operant, rulesets ...auth.RuleSet) error {
	return nil
}

func (x *stub) CanUpdate(operant auth.Operant, field string, value any, rulesets ...auth.RuleSet) error {
	return nil
}

func (x *stub) CanDelete(operant auth.Operant, rulesets ...auth.RuleSet) error {
	return nil
}

func (x *stub) CanCache(operant auth.Operant, field string, cache string, rulesets ...auth.RuleSet) error {
	return nil
}
