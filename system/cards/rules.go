package cards

import ()

type IRules interface {
	Eval(CardHolder) ([]string, error)
}
