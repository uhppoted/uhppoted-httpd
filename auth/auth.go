package auth

import (
	"github.com/google/uuid"

	"github.com/uhppoted/uhppoted-httpd/types"
)

type TokenType int

const (
	Login   TokenType = iota // 0
	Session                  // 1
)

type IAuth interface {
	Preauthenticate(loginId uuid.UUID) (string, error)
	Authorize(uid, pwd string, sessionId uuid.UUID) (string, error)
	Verify(tokenType TokenType, token string) error
	Authorized(token, resource string) (string, string, error)
	GetLoginId(token string) (*uuid.UUID, error)
	GetSessionId(token string) (*uuid.UUID, error)
}

type OpAuth interface {
	UID() string

	CanAddController(controller Operant) error
	CanUpdateController(original Operant, updated Operant) error

	CanAddCardHolder(cardHolder *types.CardHolder) error
	CanUpdateCardHolder(original, updated *types.CardHolder) error
	CanDeleteCardHolder(cardHolder *types.CardHolder) error
}

type Operant interface {
	AsRuleEntity() interface{}
}

func NewAuthProvider(config string, loginExpiry, sessionExpiry string) (IAuth, error) {
	return NewLocalAuthProvider(config, loginExpiry, sessionExpiry)
}
