package httpd

import (
	"encoding/json"
	"fmt"

	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/hyperjumptech/grule-rule-engine/engine"

	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system/cards"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type authorizator struct {
	uid   string
	role  string
	grule *ast.KnowledgeLibrary
}

type card struct {
	Name   string
	Card   uint32
	Groups []string
}

type result struct {
	Allow  bool
	Refuse bool
}

func (op *card) HasGroup(g string) bool {
	for _, p := range op.Groups {
		if p == g {
			return true
		}
	}

	return false
}

func NewAuthorizator(uid, role string, grule *ast.KnowledgeLibrary) (*authorizator, error) {
	return &authorizator{
		uid:   uid,
		role:  role,
		grule: grule,
	}, nil
}

func (a *authorizator) UID() string {
	if a != nil {
		return a.uid
	}

	return "?"
}

func (a *authorizator) CanUpdateInterface(lan auth.Operant, field string, value interface{}) error {
	msg := fmt.Errorf("Not authorized to update interface %v", lan)
	err := fmt.Errorf("Not authorized for operation %v field:%v value:%v", "update::interface", field, value)

	if a != nil && lan != nil {
		r := result{
			Allow:  false,
			Refuse: false,
		}

		m := map[string]interface{}{
			"INTERFACE": lan.AsRuleEntity(),
			"FIELD":     field,
			"VALUE":     value,
		}

		if err = a.eval("system", "update::interface", &r, m); err != nil {
			return types.Unauthorized(msg, err)
		}

		if r.Allow && !r.Refuse {
			return nil
		}

		err = fmt.Errorf("Not authorized for %s", fmt.Sprintf("update::interface %v, field:%v, value:%v", toString(lan), field, value))
	}

	return types.Unauthorized(msg, err)
}

func (a *authorizator) CanAddController(controller auth.Operant) error {
	msg := fmt.Errorf("Not authorized to add controller")
	err := fmt.Errorf("Not authorized for operation %s", "add::controller")

	if a != nil && controller != nil {
		r := result{
			Allow:  false,
			Refuse: false,
		}

		m := map[string]interface{}{
			"CONTROLLER": controller.AsRuleEntity(),
		}

		if err := a.eval("system", "add::controller", &r, m); err != nil {
			return types.Unauthorized(msg, err)
		}

		if r.Allow && !r.Refuse {
			return nil
		}

		err = fmt.Errorf("Not authorized for %s", fmt.Sprintf("add::controller %s", toString(controller)))
	}

	return types.Unauthorized(msg, err)
}

func (a *authorizator) CanUpdateController(controller auth.Operant, field string, value interface{}) error {
	msg := fmt.Errorf("Not authorized to update controller %v", controller)
	err := fmt.Errorf("Not authorized for operation %s", "update::controller")

	if a != nil && controller != nil {
		r := result{
			Allow:  false,
			Refuse: false,
		}

		m := map[string]interface{}{
			"CONTROLLER": controller.AsRuleEntity(),
			"FIELD":      field,
			"VALUE":      value,
		}

		if err = a.eval("system", "update::controller", &r, m); err != nil {
			return types.Unauthorized(msg, err)
		}

		if r.Allow && !r.Refuse {
			return nil
		}

		err = fmt.Errorf("Not authorized for %s", fmt.Sprintf("update::controller %v, field:%v, value:%v", controller, field, value))
	}

	return types.Unauthorized(msg, err)
}

func (a *authorizator) CanDeleteController(controller auth.Operant) error {
	msg := fmt.Errorf("Not authorized to delete controller")
	err := fmt.Errorf("Not authorized for operation %s", "delete::controller")

	if a != nil && controller != nil {
		r := result{
			Allow:  false,
			Refuse: false,
		}

		m := map[string]interface{}{
			"CONTROLLER": controller.AsRuleEntity(),
		}

		if err := a.eval("system", "delete::controller", &r, m); err != nil {
			return types.Unauthorized(msg, err)
		}

		if r.Allow && !r.Refuse {
			return nil
		}

		err = fmt.Errorf("Not authorized for %s", fmt.Sprintf("delete::controller %s", toString(controller)))
	}

	return types.Unauthorized(msg, err)
}

func (a *authorizator) CanAddCardHolder(ch auth.Operant) error {
	if a != nil && ch != nil {
		r := result{
			Allow:  false,
			Refuse: false,
		}

		m := map[string]interface{}{
			"CH": ch.AsRuleEntity(),
		}

		if err := a.eval("cards", "add", &r, m); err != nil {
			return err
		}

		if r.Allow && !r.Refuse {
			return nil
		}

		return fmt.Errorf("Not authorized for %s", fmt.Sprintf("add::card %v", toString(ch)))
	}

	return fmt.Errorf("Not authorized for operation %s", "add::card")
}

func (a *authorizator) CanUpdateCardHolder(original, updated auth.Operant) error {
	if a != nil && original != nil && updated != nil {
		r := result{
			Allow:  false,
			Refuse: false,
		}

		m := map[string]interface{}{
			"ORIGINAL": original.AsRuleEntity(),
			"UPDATED":  updated.AsRuleEntity(),
		}

		if err := a.eval("cards", "update", &r, m); err != nil {
			return err
		}

		if r.Allow && !r.Refuse {
			return nil
		}

		return fmt.Errorf("Not authorized for %s", fmt.Sprintf("update::card %v %v", toString(original), toString(updated)))
	}

	return fmt.Errorf("Not authorized for operation %s", "update::card")
}

func (a *authorizator) CanUpdateCard(card auth.Operant, field string, value interface{}) error {
	msg := fmt.Errorf("Not authorized to update card %v", card)
	err := fmt.Errorf("Not authorized for operation %s", "update::card")

	if a != nil && card != nil {
		r := result{
			Allow:  false,
			Refuse: false,
		}

		m := map[string]interface{}{
			"CARD":  card.AsRuleEntity(),
			"FIELD": field,
			"VALUE": value,
		}

		if err = a.eval("cards", "update::card", &r, m); err != nil {
			return types.Unauthorized(msg, err)
		}

		if r.Allow && !r.Refuse {
			return nil
		}

		err = fmt.Errorf("Not authorized for %s", fmt.Sprintf("update::card %v, field:%v, value:%v", card, field, value))
	}

	return types.Unauthorized(msg, err)
}

func (a *authorizator) CanDeleteCardHolder(ch auth.Operant) error {
	if a != nil && ch != nil {
		r := result{
			Allow:  false,
			Refuse: false,
		}

		m := map[string]interface{}{
			"CH": ch.AsRuleEntity(),
		}

		if err := a.eval("cards", "delete", &r, m); err != nil {
			return err
		}

		if r.Allow && !r.Refuse {
			return nil
		}

		return fmt.Errorf("Not authorized for %s", fmt.Sprintf("delete::card %v", toString(ch)))
	}

	return fmt.Errorf("Not authorized for operation %s", "delete::card")
}

func (a *authorizator) CanAddDoor(door auth.Operant) error {
	msg := fmt.Errorf("Not authorized to add door")
	err := fmt.Errorf("Not authorized for operation %s", "add::door")

	if a != nil && door != nil {
		r := result{
			Allow:  false,
			Refuse: false,
		}

		m := map[string]interface{}{
			"DOOR": door.AsRuleEntity(),
		}

		if err := a.eval("doors", "add::door", &r, m); err != nil {
			return types.Unauthorized(msg, err)
		}

		if r.Allow && !r.Refuse {
			return nil
		}

		err = fmt.Errorf("Not authorized for %s", fmt.Sprintf("add::door %s", toString(door)))
	}

	return types.Unauthorized(msg, err)
}

func (a *authorizator) CanUpdateDoor(door auth.Operant, field string, value interface{}) error {
	msg := fmt.Errorf("Not authorized to update door %v", door)
	err := fmt.Errorf("Not authorized for operation %s", "update::door")

	if a != nil && door != nil {
		r := result{
			Allow:  false,
			Refuse: false,
		}

		m := map[string]interface{}{
			"DOOR":  door.AsRuleEntity(),
			"FIELD": field,
			"VALUE": value,
		}

		if err = a.eval("doors", "update::door", &r, m); err != nil {
			return types.Unauthorized(msg, err)
		}

		if r.Allow && !r.Refuse {
			return nil
		}

		err = fmt.Errorf("Not authorized for %s", fmt.Sprintf("update::door %v, field:%v, value:%v", door, field, value))
	}

	return types.Unauthorized(msg, err)
}

func (a *authorizator) CanDeleteDoor(door auth.Operant) error {
	msg := fmt.Errorf("Not authorized to delete door")
	err := fmt.Errorf("Not authorized for operation %s", "delete::door")

	if a != nil && door != nil {
		r := result{
			Allow:  false,
			Refuse: false,
		}

		m := map[string]interface{}{
			"DOOR": door.AsRuleEntity(),
		}

		if err := a.eval("doors", "delete::door", &r, m); err != nil {
			return types.Unauthorized(msg, err)
		}

		if r.Allow && !r.Refuse {
			return nil
		}

		err = fmt.Errorf("Not authorized for %s", fmt.Sprintf("delete::door %s", toString(door)))
	}

	return types.Unauthorized(msg, err)
}

func (a *authorizator) eval(ruleset string, op string, r *result, m map[string]interface{}) error {
	context := ast.NewDataContext()

	if err := context.Add("OP", op); err != nil {
		return err
	}

	if err := context.Add("RESULT", r); err != nil {
		return err
	}

	for k, v := range m {
		if err := context.Add(k, v); err != nil {
			return err
		}
	}

	kb := a.grule.NewKnowledgeBaseInstance(ruleset, "0.0.0")
	enjin := engine.NewGruleEngine()
	if err := enjin.Execute(context, kb); err != nil {
		return err
	}

	return nil
}

func makeOP(ch cards.CardHolder) *card {
	cardNumber := uint32(0)
	if ch.Card != nil {
		cardNumber = uint32(*ch.Card)
	}

	groups := []string{}
	for k, v := range ch.Groups {
		if v {
			groups = append(groups, k)
		}
	}

	return &card{
		Name:   fmt.Sprintf("%v", ch.Name),
		Card:   cardNumber,
		Groups: groups,
	}
}

func toString(entity interface{}) string {
	if b, err := json.Marshal(entity); err == nil {
		return string(b)
	}

	return fmt.Sprintf("%+v", entity)
}
