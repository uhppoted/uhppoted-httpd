package auth

import (
	"github.com/google/uuid"
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
	CanDeleteController(controller Operant) error

	CanAddCardHolder(cardHolder Operant) error
	CanUpdateCardHolder(original, updated Operant) error
	CanDeleteCardHolder(cardHolder Operant) error
}

type Operant interface {
	AsRuleEntity() interface{}
}

func NewAuthProvider(config string, loginExpiry, sessionExpiry string) (IAuth, error) {
	return NewLocalAuthProvider(config, loginExpiry, sessionExpiry)
}
