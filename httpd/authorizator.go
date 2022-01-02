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

var rulesets = map[auth.RuleSet]string{
	auth.Interfaces:  "system",
	auth.Controllers: "system",
	auth.Doors:       "doors",
	auth.Cards:       "cards",
	auth.Groups:      "groups",
	auth.Events:      "events",
	auth.Logs:        "logs",
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

func (a *authorizator) CanView(r auth.RuleSet, operant auth.Operant, field string, value interface{}) error {
	msg := fmt.Errorf("Not authorized to view %v", operant)
	err := fmt.Errorf("Not authorized to view %v field:%v value:%v", operant, field, value)

	if a != nil && operant != nil {
		tag, object := operant.AsRuleEntity()
		op := fmt.Sprintf("view::%v", tag)

		m := map[string]interface{}{
			"OBJECT": object,
			"FIELD":  field,
			"VALUE":  value,
		}

		rs := result{
			Allow:  true,
			Refuse: false,
		}

		if err := a.eval(rulesets[r], op, &rs, m); err != nil {
			return types.Unauthorised(msg, err)
		}

		if !rs.Allow || rs.Refuse {
			return types.Unauthorised(msg, err)
		}
	}

	return nil
}

func (a *authorizator) CanAdd(r auth.RuleSet, operant auth.Operant) error {
	msg := fmt.Errorf("Not authorized to add %v", operant)
	err := fmt.Errorf("Not authorized to add %v", operant)

	if a != nil && operant != nil {
		tag, object := operant.AsRuleEntity()
		op := fmt.Sprintf("add::%v", tag)

		m := map[string]interface{}{
			"OBJECT": object,
		}

		rs := result{
			Allow:  false,
			Refuse: false,
		}

		if err := a.eval(rulesets[r], op, &rs, m); err != nil {
			return types.Unauthorised(msg, err)
		}

		if rs.Allow && !rs.Refuse {
			return nil
		}
	}

	return types.Unauthorised(msg, err)
}

func (a *authorizator) CanUpdate(r auth.RuleSet, operant auth.Operant, field string, value interface{}) error {
	msg := fmt.Errorf("Not authorized to update %v", operant)
	err := fmt.Errorf("Not authorized to update %v field:%v value:%v", operant, field, value)

	if a != nil && operant != nil {
		tag, object := operant.AsRuleEntity()
		op := fmt.Sprintf("update::%v", tag)

		m := map[string]interface{}{
			"OBJECT": object,
			"FIELD":  field,
			"VALUE":  value,
		}

		rs := result{
			Allow:  false,
			Refuse: false,
		}

		if err = a.eval(rulesets[r], op, &rs, m); err != nil {
			return types.Unauthorised(msg, err)
		}

		if rs.Allow && !rs.Refuse {
			return nil
		}
	}

	return types.Unauthorised(msg, err)
}

func (a *authorizator) CanDelete(r auth.RuleSet, operant auth.Operant) error {
	msg := fmt.Errorf("Not authorized to delete %v", operant)
	err := fmt.Errorf("Not authorized to delete %v", operant)

	if a != nil && operant != nil {
		tag, object := operant.AsRuleEntity()
		op := fmt.Sprintf("delete::%v", tag)

		rs := result{
			Allow:  false,
			Refuse: false,
		}

		m := map[string]interface{}{
			"OBJECT": object,
		}

		if err := a.eval(rulesets[r], op, &rs, m); err != nil {
			return types.Unauthorised(msg, err)
		}

		if rs.Allow && !rs.Refuse {
			return nil
		}
	}

	return types.Unauthorised(msg, err)
}

func (a *authorizator) CanDeleteController(controller auth.Operant) error {
	msg := fmt.Errorf("Not authorized to delete controller")
	err := fmt.Errorf("Not authorized for operation %s", "delete::controller")

	if a != nil && controller != nil {
		_, object := controller.AsRuleEntity()

		r := result{
			Allow:  false,
			Refuse: false,
		}

		m := map[string]interface{}{
			"CONTROLLER": object,
			"FIELD":      "",
			"VALUE":      "",
		}

		if err := a.eval("system", "delete::controller", &r, m); err != nil {
			return types.Unauthorised(msg, err)
		}

		if r.Allow && !r.Refuse {
			return nil
		}

		err = fmt.Errorf("Not authorized for %s", fmt.Sprintf("delete::controller %s", toString(controller)))
	}

	return types.Unauthorised(msg, err)
}

func (a *authorizator) CanDeleteCard(card auth.Operant) error {
	ruleset := "cards"
	op := "delete::card"
	msg := fmt.Errorf("Not authorized to delete card")
	_, object := card.AsRuleEntity()

	m := map[string]interface{}{
		"CARD": object,
	}

	return a.evaluate(ruleset, op, card, m, msg)
}

func (a *authorizator) CanDeleteGroup(group auth.Operant) error {
	ruleset := "groups"
	op := "delete::group"
	msg := fmt.Errorf("Not authorized to delete group")
	_, object := group.AsRuleEntity()

	m := map[string]interface{}{
		"GROUP": object,
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
			return types.Unauthorised(msg, err)
		}

		if r.Allow && !r.Refuse {
			return nil
		}

		err = fmt.Errorf("Not authorized for %s", fmt.Sprintf("%v %v", op, toString(operant)))
	}

	return types.Unauthorised(msg, err)
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
