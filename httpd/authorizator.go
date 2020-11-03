package httpd

import (
	"fmt"
	//	"regexp"
	"sort"
	"strings"

	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/hyperjumptech/grule-rule-engine/builder"
	"github.com/hyperjumptech/grule-rule-engine/engine"
	"github.com/hyperjumptech/grule-rule-engine/pkg"

	"github.com/uhppoted/uhppoted-httpd/types"
)

type authorizator struct {
	uid     string
	role    string
	rules   []string
	library *ast.KnowledgeLibrary
}

type operation struct {
	Operation string
	Name      string
	Card      int
	Groups    []string
	Allow     bool
	Refuse     bool
	op        string
}

func (op *operation) HasGroup(g string) bool {
	for _, p := range op.Groups {
		if p == g {
			return true
		}
	}

	return false
}

var rules = `
rule AddCheckCardNumber "Check the card number is greater than 6000000" {
     when
	     OP.Operation == "add" && OP.Card < 6000000
	 then
	     OP.Allow = false;
         Retract("AddCheckCardNumber");
}

rule AddCheckGryffindor "Check that Gryffindor is not ticked" {
     when
         OP.Operation == "add" && OP.HasGroup("G04")
     then
        OP.Refuse = true;
        Retract("AddCheckGryffindor");
 	 }
`

func NewAuthorizator(uid, role string) (*authorizator, error) {
	library := ast.NewKnowledgeLibrary()
	builder := builder.NewRuleBuilder(library)

	bytes := pkg.NewBytesResource([]byte(rules))
	if err := builder.BuildRuleFromResource("uhppoted", "0.0.0", bytes); err != nil {
		return nil, err
	}

	return &authorizator{
		uid:     uid,
		role:    role,
		library: library,
		rules: []string{
			`^update::card (.*?):[6-9][0-9]{6,}:(.*?):(.*?):(.*?)* (.*?):[6-9][0-9]{6,}:(.*?):(.*?):((G02|G03|G04|G05|G06|G07|G08|G09|G10)(,?))*$`,
			`^delete::card (.*?):[6-9][0-9]{6,}:(.*?):(.*?):((G02|G03|G04|G05|G06|G07|G08|G09|G10)(?:,?))*$`,
		},
	}, nil
}

func (a *authorizator) UID() string {
	if a != nil {
		return a.uid
	}

	return "?"
}

func (a *authorizator) CanAddCardHolder(ch *types.CardHolder) error {
	if a != nil && ch != nil {
		groups := []string{}
		for k, v := range ch.Groups {
			if v {
				groups = append(groups, k)
			}
		}

		f := operation{
			Operation: "add",
			Name:      fmt.Sprintf("%v", ch.Name),
			Card:      int(*ch.Card),
			Groups:    groups,
			Allow:     true,
			Refuse:     false,
			op:        fmt.Sprintf("add::card %v", cardHolderToString(ch)),
		}

		context := ast.NewDataContext()
		if err := context.Add("OP", &f); err != nil {
			return err
		}

		kb := a.library.NewKnowledgeBaseInstance("uhppoted", "0.0.0")
		enjin := engine.NewGruleEngine()
		if err := enjin.Execute(context, kb); err != nil {
			return err
		}

		if f.Allow  && !f.Refuse {
			return nil
		}

		return fmt.Errorf("Not authorized for %s", f.op)
	}

	return fmt.Errorf("Not authorized for operation %s", "add::card")
}

func (a *authorizator) CanUpdateCardHolder(original, updated *types.CardHolder) error {
	//	if a != nil && original != nil && updated != nil {
	//		op := fmt.Sprintf("update::card %v %v", cardHolderToString(original), cardHolderToString(updated))
	//
	//		for _, s := range a.rules {
	//			matched, err := regexp.Match(s, []byte(op))
	//
	//			if err != nil {
	//				return err
	//			}
	//
	//			if matched {
	//				return nil
	//			}
	//		}
	//		return fmt.Errorf("Not authorized for %s", op)
	//	}
	//
	//	return fmt.Errorf("Not authorized for operation %s", "update::card")
	return nil
}

func (a *authorizator) CanDeleteCardHolder(ch *types.CardHolder) error {
	//	if a != nil && ch != nil {
	//		op := fmt.Sprintf("delete::card %v", cardHolderToString(ch))
	//
	//		for _, s := range a.rules {
	//			matched, err := regexp.Match(s, []byte(op))
	//
	//			if err != nil {
	//				return err
	//			}
	//
	//			if matched {
	//				return nil
	//			}
	//		}
	//		return fmt.Errorf("Not authorized for %s", op)
	//	}
	//
	//	return fmt.Errorf("Not authorized for operation %s", "delete::card")
	return nil
}

func groups(ch *types.CardHolder) string {
	if ch == nil {
		return ""
	} else {
		groups := []string{}
		for k, v := range ch.Groups {
			if v {
				groups = append(groups, k)
			}
		}

		if len(groups) > 0 {
			sort.Strings(groups)
			s := groups[0]
			for _, g := range groups[1:] {
				s += fmt.Sprintf(",%v", g)
			}
			return s
		}

		return ""
	}
}

func cardHolderToString(ch *types.CardHolder) string {
	if ch != nil {
		name := strings.ReplaceAll(fmt.Sprintf("%v", ch.Name), ":", "")
		card := strings.ReplaceAll(fmt.Sprintf("%v", ch.Card), ":", "")
		from := strings.ReplaceAll(fmt.Sprintf("%v", ch.From), ":", "")
		to := strings.ReplaceAll(fmt.Sprintf("%v", ch.To), ":", "")

		s := fmt.Sprintf("%v:%v:%v:%v:", name, card, from, to)

		groups := []string{}
		for k, v := range ch.Groups {
			if v {
				groups = append(groups, k)
			}
		}

		if len(groups) > 0 {
			sort.Strings(groups)
			s += strings.ReplaceAll(groups[0], ":", "")
			for _, g := range groups[1:] {
				s += "," + strings.ReplaceAll(g, ":", "")
			}
		}

		return strings.ReplaceAll(s, " ", "")
	}

	return ""
}
