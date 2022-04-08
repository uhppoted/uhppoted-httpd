package grule

import (
	"regexp"
	"strings"

	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/hyperjumptech/grule-rule-engine/engine"

	"github.com/uhppoted/uhppoted-httpd/system/cards"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	"github.com/uhppoted/uhppoted-httpd/system/doors"
)

type Rules interface {
	Eval(cards.Card, doors.Doors) ([]doors.Door, []doors.Door, error)
}

type rules struct {
	grule *ast.KnowledgeLibrary
}

type permissions struct {
	allowed   []string
	forbidden []string
}

type query struct {
}

func (q query) HasGroup(groups []string, group string) bool {
	v := clean(group)
	for _, g := range groups {
		if clean(g) == v {
			return true
		}
	}

	return false
}

func (p *permissions) Allow(door string) {
	p.allowed = append(p.allowed, door)
}

func (p *permissions) Forbid(door string) {
	p.forbidden = append(p.forbidden, door)
}

func NewGrule(library *ast.KnowledgeLibrary) (*rules, error) {
	return &rules{
		grule: library,
	}, nil
}

func (r *rules) Eval(c cards.Card, dd doors.Doors) ([]doors.Door, []doors.Door, error) {
	if r != nil {
		p := permissions{
			allowed:   []string{},
			forbidden: []string{},
		}

		context := ast.NewDataContext()

		_, e := c.AsRuleEntity()
		if err := context.Add("CARD", e); err != nil {
			return nil, nil, err
		}

		if err := context.Add("DOORS", &p); err != nil {
			return nil, nil, err
		}

		if err := context.Add("QUERY", &query{}); err != nil {
			return nil, nil, err
		}

		kb := r.grule.NewKnowledgeBaseInstance("acl", "0.0.0")
		enjin := engine.NewGruleEngine()
		if err := enjin.Execute(context, kb); err != nil {
			return nil, nil, err
		}

		list := map[schema.OID]int{}
		for _, a := range p.allowed {
			if d, ok := dd.ByName(a); ok {
				list[d.OID] = 1
			}
		}

		for _, f := range p.forbidden {
			if d, ok := dd.ByName(f); ok {
				list[d.OID] = -1
			}
		}

		allowed := []doors.Door{}
		forbidden := []doors.Door{}
		for k, v := range list {
			if d, ok := dd.Door(k); ok {
				switch v {
				case 1:
					allowed = append(allowed, d)
				case -1:
					forbidden = append(forbidden, d)
				}
			}
		}

		return allowed, forbidden, nil
	}

	return nil, nil, nil
}

func clean(s string) string {
	return strings.ToLower(regexp.MustCompile(`\s+`).ReplaceAllString(s, ""))
}
