package httpd

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	//	"github.com/hyperjumptech/grule-rule-engine/ast"
	//	"github.com/hyperjumptech/grule-rule-engine/builder"
	//	"github.com/hyperjumptech/grule-rule-engine/engine"
	//	"github.com/hyperjumptech/grule-rule-engine/pkg"

	"github.com/uhppoted/uhppoted-httpd/types"
)

type authorizator struct {
	uid   string
	role  string
	rules []string
}

type operation struct {
	Operation  string
	Name       string
	Card       string
	Authorised bool
	op         string
}

func (a *authorizator) UID() string {
	if a != nil {
		return a.uid
	}

	return "?"
}

func (a *authorizator) CanAddCardHolder(ch *types.CardHolder) error {
	//	if a != nil && ch != nil {
	//
	//		f := operation{
	//			Operation:  "add",
	//			Name:       fmt.Sprintf("%v", ch.Name),
	//			Card:       fmt.Sprintf("%v", ch.Card),
	//			Authorised: false,
	//			op:         fmt.Sprintf("add::card %v", cardHolderToString(ch)),
	//		}
	//
	//		context := ast.NewDataContext()
	//		err := context.Add("OP", f)
	//		if err != nil {
	//			return err
	//		}
	//
	//		library := ast.NewKnowledgeLibrary()
	//		rules := builder.NewRuleBuilder(library)
	//		rule := `
	//rule CheckCardNumber "Check the card number" {
	//    when
	//        OP.Operation == "add" && OP.Card == "6000005"
	//    then
	//        OP.Authorised = true;
	//        Retract("CheckCardNumber");
	//}
	//`
	//
	//		bytes := pkg.NewBytesResource([]byte(rule))
	//		if err := rules.BuildRuleFromResource("uhppoted", "0.0.0", bytes); err != nil {
	//			return err
	//		}
	//
	//		kb := library.NewKnowledgeBaseInstance("uhppoted", "0.0.0")
	//		enjin := engine.NewGruleEngine()
	//		if err := enjin.Execute(context, kb); err != nil {
	//			return err
	//		}
	//
	//		if f.Authorised {
	//			return nil
	//		}
	//
	//		return fmt.Errorf("Not authorized for %s", f.op)
	//	}

	if a != nil && ch != nil {
		op := fmt.Sprintf("add::card %v", cardHolderToString(ch))

		for _, s := range a.rules {
			matched, err := regexp.Match(s, []byte(op))

			if err != nil {
				return err
			}

			if matched {
				return nil
			}
		}
		return fmt.Errorf("not authorised for %s", op)
	}

	return fmt.Errorf("Not authorized for operation %s", "add::card")
}

func (a *authorizator) CanUpdateCardHolder(original, updated *types.CardHolder) error {
	if a != nil && original != nil && updated != nil {
		op := fmt.Sprintf("update::card %v %v", cardHolderToString(original), cardHolderToString(updated))

		for _, s := range a.rules {
			matched, err := regexp.Match(s, []byte(op))

			if err != nil {
				return err
			}

			if matched {
				return nil
			}
		}
		return fmt.Errorf("Not authorized for %s", op)
	}

	return fmt.Errorf("Not authorized for operation %s", "update::card")
}

func (a *authorizator) CanDeleteCardHolder(ch *types.CardHolder) error {
	if a != nil && ch != nil {
		op := fmt.Sprintf("delete::card %v", cardHolderToString(ch))

		for _, s := range a.rules {
			matched, err := regexp.Match(s, []byte(op))

			if err != nil {
				return err
			}

			if matched {
				return nil
			}
		}
		return fmt.Errorf("Not authorized for %s", op)
	}

	return fmt.Errorf("Not authorized for operation %s", "delete::card")
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
