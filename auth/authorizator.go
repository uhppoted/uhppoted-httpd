package auth

import (
	"embed"
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/hyperjumptech/grule-rule-engine/builder"
	"github.com/hyperjumptech/grule-rule-engine/engine"
	"github.com/hyperjumptech/grule-rule-engine/pkg"

	"github.com/uhppoted/uhppoted-httpd/log"
)

type TAuthable interface {
	AsRuleEntity() (string, any)
	CacheKey() string
}

type authorizator struct {
	uid  string
	role string
}

type result struct {
	Allow  bool
	Refuse bool
}

type ruleset struct {
	kb      *ast.KnowledgeLibrary
	file    string
	touched time.Time
}

var grules = struct {
	ruleset map[RuleSet]ruleset
	roles   struct {
		admin string
	}

	sync.RWMutex
}{
	ruleset: map[RuleSet]ruleset{},

	roles: struct {
		admin string
	}{
		admin: "admin",
	},
}

//go:embed grules
var GRULES embed.FS

func Init(rules map[RuleSet]string, adminRole string) error {
	grules.Lock()
	defer grules.Unlock()

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
			return fmt.Errorf("error loading %v auth ruleset (%v)", "interfaces", err)
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

	grules.roles.admin = adminRole

	return nil
}

func CanView[T TAuthable](a OpAuth, u T, field string, value any, rulesets ...RuleSet) error {
	if !isNil(a) {
		return a.CanView(u, field, value, rulesets...)
	}

	return nil
}

func CanAdd[T TAuthable](a OpAuth, u T, rulesets ...RuleSet) error {
	if !isNil(a) {
		return a.CanAdd(u, rulesets...)
	}

	return nil
}

func CanUpdate[T TAuthable](a OpAuth, u T, field string, value any, rulesets ...RuleSet) error {
	if !isNil(a) {
		return a.CanUpdate(u, field, value, rulesets...)
	}

	return nil
}

func CanDelete[T TAuthable](a OpAuth, u T, rulesets ...RuleSet) error {
	if !isNil(a) {
		return a.CanDelete(u, rulesets...)
	}

	return nil
}

func isNil(v any) bool {
	if v == nil {
		return true
	}

	switch reflect.TypeOf(v).Kind() {
	case reflect.Ptr,
		reflect.Map,
		reflect.Array,
		reflect.Chan,
		reflect.Slice:
		return reflect.ValueOf(v).IsNil()
	}

	return false
}

func (a *authorizator) CanView(operant Operant, field string, value any, rulesets ...RuleSet) error {
	f := func() (result, error) {
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
				return rs, ErrUnauthorised
			}
		}

		return rs, nil
	}

	if a != nil && operant != nil {
		if rs, err := cacheCanView(a.uid, a.role, operant, field, f); err != nil {
			return ErrUnauthorised
		} else if !rs.Allow || rs.Refuse {
			return ErrUnauthorised
		}
	}

	return nil
}

func (a *authorizator) CanAdd(operant Operant, rulesets ...RuleSet) error {
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
				return ErrUnauthorised
			}
		}

		if rs.Allow && !rs.Refuse {
			return nil
		}
	}

	return ErrUnauthorised
}

func (a *authorizator) CanUpdate(operant Operant, field string, value interface{}, rulesets ...RuleSet) error {
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
			if err := a.eval(r, op, &rs, m); err != nil {
				return ErrUnauthorised
			}
		}

		if rs.Allow && !rs.Refuse {
			return nil
		}
	}

	return ErrUnauthorised
}

func (a *authorizator) CanDelete(operant Operant, rulesets ...RuleSet) error {
	if !isNil(operant) {
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
				return ErrUnauthorised
			}
		}

		if rs.Allow && !rs.Refuse {
			return nil
		}
	}

	return ErrUnauthorised
}

func (a *authorizator) eval(ruleset RuleSet, op string, r *result, m map[string]interface{}) error {
	context := ast.NewDataContext()
	tag := fmt.Sprintf("%v", ruleset)

	if err := context.Add("UID", a.uid); err != nil {
		return err
	}

	if err := context.Add("ROLE", a.role); err != nil {
		return err
	}

	if err := context.Add("OP", op); err != nil {
		return err
	}

	if err := context.Add("ADMIN", grules.roles.admin); err != nil {
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
	} else if kbi, err := kb.NewKnowledgeBaseInstance(tag, "0.0.0"); err != nil {
		return err
	} else {
		enjin := engine.NewGruleEngine()
		if err := enjin.Execute(context, kbi); err != nil {
			return err
		}
	}

	return nil
}

func getKB(r RuleSet) (*ast.KnowledgeLibrary, error) {
	grules.RLock()
	v, ok := grules.ruleset[r]
	grules.RUnlock()

	if !ok || (v.kb == nil && v.file == "") {
		return nil, fmt.Errorf("no rules knowledgebase for ruleset '%v'", r)
	}

	if v.file == "" {
		return v.kb, nil
	}

	var touched time.Time
	var tag = fmt.Sprintf("%v", r)

	if info, err := os.Stat(v.file); err != nil {
		return nil, fmt.Errorf("error loading %v auth ruleset (%v)", tag, err)
	} else {
		touched = info.ModTime()
	}

	if v.kb != nil && !v.touched.Before(touched) {
		return v.kb, nil
	}

	kb := ast.NewKnowledgeLibrary()
	if err := builder.NewRuleBuilder(kb).BuildRuleFromResource(tag, "0.0.0", pkg.NewFileResource(v.file)); err != nil {
		return nil, fmt.Errorf("error loading %v auth ruleset (%v)", tag, err)
	}

	grules.Lock()
	defer grules.Unlock()

	grules.ruleset[r] = ruleset{
		kb:      kb,
		file:    v.file,
		touched: touched,
	}

	cacheClear()

	infof("AUTH", "loaded '%v' grule file from %v", tag, v.file)

	return kb, nil
}

func infof(tag string, format string, args ...any) {
	if tag == "" {
		log.Infof("%v", args...)
	} else {
		log.Infof(fmt.Sprintf("%-8v %v", tag, format), args...)
	}
}
