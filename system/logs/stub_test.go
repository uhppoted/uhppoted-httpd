package logs

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

func (x *stub) CanUpdateInterface(lan auth.Operant, field string, value interface{}) error {
	return nil
}

func (x *stub) CanAddController(controller auth.Operant) error {
	return nil
}

func (x *stub) CanUpdateController(controller auth.Operant, field string, value interface{}) error {
	return nil
}

func (x *stub) CanDeleteController(controller auth.Operant) error {
	return nil
}

func (x *stub) CanAddCard(card auth.Operant) error {
	return nil
}

func (x *stub) CanUpdateCard(card auth.Operant, field string, value interface{}) error {
	return nil
}

func (x *stub) CanDeleteCard(card auth.Operant) error {
	return nil
}

func (x *stub) CanAddDoor(door auth.Operant) error {
	return nil
}

func (x *stub) CanUpdateDoor(door auth.Operant, field string, value interface{}) error {
	return nil
}

func (x *stub) CanDeleteDoor(door auth.Operant) error {
	return nil
}

func (x *stub) CanAddGroup(group auth.Operant) error {
	return nil
}

func (x *stub) CanUpdateGroup(group auth.Operant, field string, value interface{}) error {
	return nil
}

func (x *stub) CanDeleteGroup(group auth.Operant) error {
	return nil
}
