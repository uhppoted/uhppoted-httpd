package httpd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/hyperjumptech/grule-rule-engine/builder"
	"github.com/hyperjumptech/grule-rule-engine/engine"
	"github.com/hyperjumptech/grule-rule-engine/pkg"

	"github.com/uhppoted/uhppoted-httpd/auth"
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

var grules = map[string]struct {
	kb      *ast.KnowledgeLibrary
	touched time.Time
}{}

func NewAuthorizator(uid, role, tag, rules string) (*authorizator, error) {
	var kb *ast.KnowledgeLibrary
	var touched time.Time

	if info, err := os.Stat(rules); err != nil {
		return nil, fmt.Errorf("Error loading %v auth ruleset (%v)", tag, err)
	} else {
		touched = info.ModTime()
	}

	if v, ok := grules[rules]; ok && v.kb != nil && !v.touched.Before(touched) {
		kb = v.kb
	} else {
		kb = ast.NewKnowledgeLibrary()
		if err := builder.NewRuleBuilder(kb).BuildRuleFromResource(tag, "0.0.0", pkg.NewFileResource(rules)); err != nil {
			return nil, fmt.Errorf("Error loading %v auth ruleset (%v)", tag, err)
		}

		grules[rules] = struct {
			kb      *ast.KnowledgeLibrary
			touched time.Time
		}{
			kb:      kb,
			touched: touched,
		}

		log.Printf("INFO  loaded '%v' grule file from %v", tag, rules)
	}

	return &authorizator{
		uid:   uid,
		role:  role,
		grule: kb,
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

func (a *authorizator) CanAddCard(card auth.Operant) error {
	msg := fmt.Errorf("Not authorized to add card")
	err := fmt.Errorf("Not authorized for operation %s", "add::card")

	if a != nil && card != nil {
		r := result{
			Allow:  false,
			Refuse: false,
		}

		m := map[string]interface{}{
			"CARD": card.AsRuleEntity(),
		}

		if err := a.eval("cards", "add::card", &r, m); err != nil {
			return types.Unauthorized(msg, err)
		}

		if r.Allow && !r.Refuse {
			return nil
		}

		err = fmt.Errorf("Not authorized for %s", fmt.Sprintf("add::card %s", toString(card)))
	}

	return types.Unauthorized(msg, err)
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

func (a *authorizator) CanDeleteCard(card auth.Operant) error {
	ruleset := "cards"
	op := "delete::card"
	msg := fmt.Errorf("Not authorized to delete card")

	m := map[string]interface{}{
		"CARD": card.AsRuleEntity(),
	}

	return a.evaluate(ruleset, op, card, m, msg)
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

func (a *authorizator) CanAddGroup(group auth.Operant) error {
	ruleset := "groups"
	op := "add::group"
	msg := fmt.Errorf("Not authorized to add group")

	m := map[string]interface{}{
		"GROUP": group.AsRuleEntity(),
	}

	return a.evaluate(ruleset, op, group, m, msg)
}

func (a *authorizator) CanUpdateGroup(group auth.Operant, field string, value interface{}) error {
	ruleset := "groups"
	op := "update::group"
	msg := fmt.Errorf("Not authorized to update group %v", group)
	err := fmt.Errorf("Not authorized for operation %s", op)

	if a != nil && group != nil {
		r := result{
			Allow:  false,
			Refuse: false,
		}

		m := map[string]interface{}{
			"GROUP": group.AsRuleEntity(),
			"FIELD": field,
			"VALUE": value,
		}

		if err = a.eval(ruleset, op, &r, m); err != nil {
			return types.Unauthorized(msg, err)
		}

		if r.Allow && !r.Refuse {
			return nil
		}

		err = fmt.Errorf("Not authorized for %s", fmt.Sprintf("%v %v, field:%v, value:%v", op, group, field, value))
	}

	return types.Unauthorized(msg, err)
}

func (a *authorizator) CanDeleteGroup(group auth.Operant) error {
	ruleset := "groups"
	op := "delete::group"
	msg := fmt.Errorf("Not authorized to delete group")

	m := map[string]interface{}{
		"GROUP": group.AsRuleEntity(),
	}

	return a.evaluate(ruleset, op, group, m, msg)
}

func (a *authorizator) evaluate(ruleset, op string, operant auth.Operant, m map[string]interface{}, msg error) error {
	err := fmt.Errorf("Not authorized for operation %s", op)

	if a != nil && operant != nil {
		r := result{
			Allow:  false,
			Refuse: false,
		}

		if err := a.eval(ruleset, op, &r, m); err != nil {
			return types.Unauthorized(msg, err)
		}

		if r.Allow && !r.Refuse {
			return nil
		}

		err = fmt.Errorf("Not authorized for %s", fmt.Sprintf("%v %v", op, toString(operant)))
	}

	return types.Unauthorized(msg, err)
}

func (a *authorizator) eval(ruleset string, op string, r *result, m map[string]interface{}) error {
	context := ast.NewDataContext()

	if err := context.Add("UID", a.uid); err != nil {
		return err
	}

	if err := context.Add("ROLE", a.role); err != nil {
		return err
	}

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

func toString(entity interface{}) string {
	if b, err := json.Marshal(entity); err == nil {
		return string(b)
	}

	return fmt.Sprintf("%+v", entity)
}
