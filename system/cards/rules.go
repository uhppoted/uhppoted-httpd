package db

import (
	"github.com/uhppoted/uhppoted-httpd/types"
)

type IRules interface {
	Eval(types.CardHolder) ([]string, error)
}
