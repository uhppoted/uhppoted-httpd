package httpd

import (
	"fmt"
	"regexp"

	"github.com/uhppoted/uhppoted-httpd/types"
)

type authorizator struct {
	uid   string
	role  string
	rules []string
}

func (a *authorizator) UID() string {
	if a != nil {
		return a.uid
	}

	return "?"
}

func (a *authorizator) CanAddCardHolder(ch *types.CardHolder) error {
	if a != nil && ch != nil {
		op := fmt.Sprintf("add::card::%v:%v", ch.Name, ch.Card)

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

	return fmt.Errorf("not authorised for %s", "add::card")
}

func (a *authorizator) CanUpdateCardHolder(original, updated *types.CardHolder) error {
	if original != nil && updated != nil {
		if original.Name.Equals(updated.Name) && original.Card.Equals(updated.Card) {
			return nil
		}
	}

	return fmt.Errorf("Nope, no can do, sorry compadre")
}

func (a *authorizator) CanDeleteCardHolder(ch *types.CardHolder) error {
	if ch != nil {
		if ch.Card != nil && *ch.Card >= 6000000 {
			return nil
		}

	}

	return fmt.Errorf("Nope, no can do, sorry compadre")
}
