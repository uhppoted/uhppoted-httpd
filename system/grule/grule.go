package grule

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/hyperjumptech/grule-rule-engine/engine"

	"github.com/uhppoted/uhppoted-httpd/system/cards"
	"github.com/uhppoted/uhppoted-httpd/system/catalog/schema"
	"github.com/uhppoted/uhppoted-httpd/system/doors"
	"github.com/uhppoted/uhppoted-httpd/system/groups"
)

type Rules interface {
	Eval(cards.Card, groups.Groups, doors.Doors) ([]doors.Door, []doors.Door, error)
}

type rules struct {
	grule *ast.KnowledgeLibrary
}

type card struct {
	name   string
	card   uint32
	groups []string
}

type permissions struct {
	allowed   []string
	forbidden []string
}

func (c *card) String() string {
	return fmt.Sprintf("%v (%v)", c.name, c.card)
}

func (c *card) Has(field string, value interface{}) bool {
	f := strings.ToLower(strings.TrimSpace(field))

	var v string
	if value != nil {
		v = clean(fmt.Sprintf("%v", value))
	} else {
		v = ""
	}

	switch f {
	case "name":
		return clean(c.name) == v

	case "group":
		for _, g := range c.groups {
			if v == clean(g) {
				return true
			}
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

func (r *rules) Eval(ch cards.Card, gg groups.Groups, dd doors.Doors) ([]doors.Door, []doors.Door, error) {
	if r != nil {
		p := permissions{
			allowed:   []string{},
			forbidden: []string{},
		}

		context := ast.NewDataContext()

		if err := context.Add("CARD", makeCard(ch, gg)); err != nil {
			return nil, nil, err
		}

		if err := context.Add("DOORS", &p); err != nil {
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

func makeCard(c cards.Card, gg groups.Groups) *card {
	groups := []string{}
	for k, v := range c.Groups() {
		if v {
			if g, ok := gg.Group(k); ok {
				groups = append(groups, g.Name)
			}
		}
	}

	return &card{
		name:   c.Name(),
		card:   c.CardNumber(),
		groups: groups,
	}
}

func clean(s string) string {
	return strings.ToLower(regexp.MustCompile(`\s+`).ReplaceAllString(s, ""))
}
