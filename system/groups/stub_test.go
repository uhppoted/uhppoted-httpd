package groups

import (
	"github.com/uhppoted/uhppoted-httpd/auth"
)

type stub struct {
	canView func(auth.RuleSet, auth.Operant, string, interface{}) error
}

func (x *stub) UID() string {
	return "stub"
}

func (x *stub) CanView(ruleset auth.RuleSet, object auth.Operant, field string, value interface{}) error {
	if x.canView != nil {
		return x.canView(ruleset, object, field, value)
	}

	return nil
}

func (x *stub) CanAdd(ruleset auth.RuleSet, operant auth.Operant) error {
	return nil
}

func (x *stub) CanUpdate(ruleset auth.RuleSet, operant auth.Operant, field string, value interface{}) error {
	return nil
}

func (x *stub) CanDeleteController(controller auth.Operant) error {
	return nil
}

func (x *stub) CanDeleteCard(card auth.Operant) error {
	return nil
}

func (x *stub) CanDeleteDoor(door auth.Operant) error {
	return nil
}

func (x *stub) CanDeleteGroup(group auth.Operant) error {
	return nil
}
