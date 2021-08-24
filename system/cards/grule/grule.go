package grule

import (
	"fmt"

	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/hyperjumptech/grule-rule-engine/engine"

	"github.com/uhppoted/uhppoted-httpd/system/cards"
)

type rules struct {
	grule *ast.KnowledgeLibrary
}

type card struct {
	Name   string
	Card   uint32
	Groups []string
}

type doors struct {
	allowed   []string
	forbidden []string
}

func (c *card) HasGroup(g string) bool {
	for _, p := range c.Groups {
		if p == g {
			return true
		}
	}

	return false
}

func (d *doors) Allow(door string) {
	for _, p := range d.allowed {
		if p == door {
			return
		}
	}

	d.allowed = append(d.allowed, door)
}

func (d *doors) Forbid(door string) {
	for _, p := range d.forbidden {
		if p == door {
			return
		}
	}

	d.forbidden = append(d.forbidden, door)
}

func NewGrule(library *ast.KnowledgeLibrary) (*rules, error) {
	return &rules{
		grule: library,
	}, nil
}

func (r *rules) Eval(ch cards.CardHolder) ([]string, error) {
	if r != nil {
		dd := doors{
			allowed:   []string{},
			forbidden: []string{},
		}

		context := ast.NewDataContext()

		if err := context.Add("CH", makeCH(ch)); err != nil {
			return nil, err
		}

		if err := context.Add("DOORS", &dd); err != nil {
			return nil, err
		}

		kb := r.grule.NewKnowledgeBaseInstance("acl", "0.0.0")
		enjin := engine.NewGruleEngine()
		if err := enjin.Execute(context, kb); err != nil {
			return nil, err
		}

		permissions := []string{}

	loop:
		for _, p := range dd.allowed {
			for _, q := range dd.forbidden {
				if q == p {
					continue loop
				}
			}
			permissions = append(permissions, p)
		}

		return permissions, nil
	}

	return nil, nil
}

func makeCH(ch cards.CardHolder) *card {
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
