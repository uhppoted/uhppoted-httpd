package db

import (
	"github.com/uhppoted/uhppoted-httpd/types"
)

type IAuth interface {
	CanAddCardHolder(cardHolder *types.CardHolder) error
	CanUpdateCardHolder(original, updated *types.CardHolder) error
	CanDeleteCardHolder(cardHolder *types.CardHolder) error
}
