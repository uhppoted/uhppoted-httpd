package grule

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/hyperjumptech/grule-rule-engine/engine"

	"github.com/uhppoted/uhppoted-httpd/system/cards"
	"github.com/uhppoted/uhppoted-httpd/system/catalog"
	"github.com/uhppoted/uhppoted-httpd/system/doors"
	"github.com/uhppoted/uhppoted-httpd/system/groups"
)

type Rules interface {
	Eval(cards.CardHolder, groups.Groups, doors.Doors) ([]doors.Door, []doors.Door, error)
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

func (r *rules) Eval(ch cards.CardHolder, gg groups.Groups, dd doors.Doors) ([]doors.Door, []doors.Door, error) {
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

		list := map[catalog.OID]int{}
		for _, a := range p.allowed {
			for _, d := range dd.Doors {
				if clean(a) == clean(d.Name) {
					list[d.OID] = 1
					break
				}
			}
		}

		for _, f := range p.forbidden {
			for _, d := range dd.Doors {
				if clean(f) == clean(d.Name) {
					list[d.OID] = -1
					break
				}
			}
		}

		allowed := []doors.Door{}
		forbidden := []doors.Door{}
		for k, v := range list {
			if d, ok := dd.Doors[k]; ok {
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

func makeCard(ch cards.CardHolder, groups groups.Groups) *card {
	cardNumber := uint32(0)
	if ch.Card != nil {
		cardNumber = uint32(*ch.Card)
	}

	c := card{
		name:   fmt.Sprintf("%v", ch.Name),
		card:   cardNumber,
		groups: []string{},
	}

	for k, v := range ch.Groups {
		if v {
			if g, ok := groups.Groups[k]; ok {
				c.groups = append(c.groups, g.Name)
			}
		}
	}

	return &c
}

func clean(s string) string {
	return strings.ToLower(regexp.MustCompile(`\s+`).ReplaceAllString(s, ""))
}
