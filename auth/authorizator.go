package auth

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/hyperjumptech/grule-rule-engine/builder"
	"github.com/hyperjumptech/grule-rule-engine/engine"
	"github.com/hyperjumptech/grule-rule-engine/pkg"

	"github.com/uhppoted/uhppoted-httpd/types"
)

type authorizator struct {
	uid  string
	role string
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

type ruleset struct {
	kb      *ast.KnowledgeLibrary
	file    string
	touched time.Time
}

var grules = struct {
	ruleset map[RuleSet]ruleset
	guard   sync.RWMutex
}{
	ruleset: map[RuleSet]ruleset{},
}

//go:embed grules
var GRULES embed.FS

func Init(rules map[RuleSet]string) error {
	grules.guard.Lock()
	defer grules.guard.Unlock()

	// ... initialise from embedded grules files
	var touched time.Time

	resources := map[RuleSet]struct {
		tag  string
		file string
	}{
		Interfaces:  {"interfaces", "grules/interfaces.grl"},
		Controllers: {"controllers", "grules/controllers.grl"},
		Doors:       {"doors", "grules/doors.grl"},
		Cards:       {"cards", "grules/cards.grl"},
		Groups:      {"groups", "grules/groups.grl"},
		Events:      {"events", "grules/events.grl"},
		Logs:        {"logs", "grules/logs.grl"},
		Users:       {"users", "grules/users.grl"},
	}

	for k, v := range resources {
		kb := ast.NewKnowledgeLibrary()
		resource := pkg.NewEmbeddedResource(GRULES, v.file)
		if err := builder.NewRuleBuilder(kb).BuildRuleFromResource(v.tag, "0.0.0", resource); err != nil {
			return fmt.Errorf("Error loading %v auth ruleset (%v)\n", "interfaces", err)
		} else {
			grules.ruleset[k] = ruleset{
				kb:      kb,
				touched: touched,
			}
		}
	}

	for k, v := range rules {
		if f := strings.TrimSpace(v); f != "" {
			grules.ruleset[k] = struct {
				kb      *ast.KnowledgeLibrary
				file    string
				touched time.Time
			}{
				file: f,
			}
		}
	}

	return nil
}

func NewAuthorizator(uid, role string) OpAuth {
	return &authorizator{
		uid:  uid,
		role: role,
	}
}

func (a *authorizator) UID() string {
	if a != nil {
		return a.uid
	}

	return "?"
}

func (a *authorizator) CanView(operant Operant, field string, value interface{}, rulesets ...RuleSet) error {
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

		for _, r := range rulesets {
			if err := a.eval(r, op, &rs, m); err != nil {
				return types.Unauthorised(msg, err)
			}
		}

		if !rs.Allow || rs.Refuse {
			return types.Unauthorised(msg, err)
		}
	}

	return nil
}

func (a *authorizator) CanAdd(operant Operant, rulesets ...RuleSet) error {
	msg := fmt.Errorf("Not authorized to add %v", operant)
	err := fmt.Errorf("Not authorized to add %v", operant)

	if a != nil && operant != nil {
		tag, object := operant.AsRuleEntity()
		op := fmt.Sprintf("add::%v", tag)

		m := map[string]interface{}{
			"OBJECT": object,
			"FIELD":  "",
		}

		rs := result{
			Allow:  false,
			Refuse: false,
		}

		for _, r := range rulesets {
			if err := a.eval(r, op, &rs, m); err != nil {
				return types.Unauthorised(msg, err)
			}
		}

		if rs.Allow && !rs.Refuse {
			return nil
		}
	}

	return types.Unauthorised(msg, err)
}

func (a *authorizator) CanUpdate(operant Operant, field string, value interface{}, rulesets ...RuleSet) error {
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

		for _, r := range rulesets {
			if err = a.eval(r, op, &rs, m); err != nil {
				return types.Unauthorised(msg, err)
			}
		}

		if rs.Allow && !rs.Refuse {
			return nil
		}
	}

	return types.Unauthorised(msg, err)
}

func (a *authorizator) CanDelete(operant Operant, rulesets ...RuleSet) error {
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
			"FIELD":  "",
		}

		for _, r := range rulesets {
			if err := a.eval(r, op, &rs, m); err != nil {
				return types.Unauthorised(msg, err)
			}
		}

		if rs.Allow && !rs.Refuse {
			return nil
		}
	}

	return types.Unauthorised(msg, err)
}

func (a *authorizator) evaluate(ruleset RuleSet, op string, operant Operant, m map[string]interface{}, msg error) error {
	err := fmt.Errorf("Not authorized for operation %s", op)

	if a != nil && operant != nil {
		rs := result{
			Allow:  false,
			Refuse: false,
		}

		if err := a.eval(ruleset, op, &rs, m); err != nil {
			return types.Unauthorised(msg, err)
		}

		if rs.Allow && !rs.Refuse {
			return nil
		}

		err = fmt.Errorf("Not authorized for %s", fmt.Sprintf("%v %v", op, toString(operant)))
	}

	return types.Unauthorised(msg, err)
}

func (a *authorizator) eval(ruleset RuleSet, op string, r *result, m map[string]interface{}) error {
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

	if kb, err := getKB(ruleset); err != nil {
		return err
	} else {
		tag := fmt.Sprintf("%v", ruleset)
		kbi := kb.NewKnowledgeBaseInstance(tag, "0.0.0")
		enjin := engine.NewGruleEngine()
		if err := enjin.Execute(context, kbi); err != nil {
			return err
		}
	}

	return nil
}

func getKB(r RuleSet) (*ast.KnowledgeLibrary, error) {
	grules.guard.RLock()
	v, ok := grules.ruleset[r]
	grules.guard.RUnlock()

	if !ok || (v.kb == nil && v.file == "") {
		return nil, fmt.Errorf("No rules knowledgebase for ruleset '%v'", r)
	}

	if v.file == "" {
		return v.kb, nil
	}

	var touched time.Time
	var tag = fmt.Sprintf("%v", r)

	if info, err := os.Stat(v.file); err != nil {
		return nil, fmt.Errorf("Error loading %v auth ruleset (%v)", tag, err)
	} else {
		touched = info.ModTime()
	}

	if v.kb != nil && !v.touched.Before(touched) {
		return v.kb, nil
	}

	kb := ast.NewKnowledgeLibrary()
	if err := builder.NewRuleBuilder(kb).BuildRuleFromResource(tag, "0.0.0", pkg.NewFileResource(v.file)); err != nil {
		return nil, fmt.Errorf("Error loading %v auth ruleset (%v)", tag, err)
	}

	grules.guard.Lock()
	defer grules.guard.Unlock()

	grules.ruleset[r] = ruleset{
		kb:      kb,
		file:    v.file,
		touched: touched,
	}

	log.Printf("INFO  loaded '%v' grule file from %v", tag, v.file)

	return kb, nil
}

func toString(entity interface{}) string {
	if b, err := json.Marshal(entity); err == nil {
		return string(b)
	}

	return fmt.Sprintf("%+v", entity)
}
