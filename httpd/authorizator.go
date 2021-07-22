package httpd

import (
	"encoding/json"
	"fmt"

	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/hyperjumptech/grule-rule-engine/engine"

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

		if err := a.eval("system", "add", &r, m); err != nil {
			return types.Unauthorized(msg, err)
		}

		if r.Allow && !r.Refuse {
			return nil
		}

		err = fmt.Errorf("Not authorized for %s", fmt.Sprintf("add::controller %s", toString(controller)))
	}

	return types.Unauthorized(msg, err)
}

func (a *authorizator) CanUpdateController(original, updated auth.Operant) error {
	msg := fmt.Errorf("Not authorized to update controller")
	err := fmt.Errorf("Not authorized for operation %s", "update::controller")

	if a != nil && original != nil && updated != nil {
		r := result{
			Allow:  false,
			Refuse: false,
		}

		m := map[string]interface{}{
			"ORIGINAL": original.AsRuleEntity(),
			"UPDATED":  updated.AsRuleEntity(),
		}

		if err = a.eval("system", "update", &r, m); err != nil {
			return types.Unauthorized(msg, err)
		}

		if r.Allow && !r.Refuse {
			return nil
		}

		err = fmt.Errorf("Not authorized for %s", fmt.Sprintf("update::controller %s %s", toString(original), toString(updated)))
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

		if err := a.eval("system", "delete", &r, m); err != nil {
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

func makeOP(ch types.CardHolder) *card {
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
