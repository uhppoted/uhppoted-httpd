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

	CanUpdateInterface(iface Operant, field string, value interface{}) error

	CanAddController(controller Operant) error
	CanUpdateController(controller Operant, field string, value interface{}) error
	CanDeleteController(controller Operant) error

	CanAddCardHolder(cardHolder Operant) error
	CanUpdateCardHolder(original, updated Operant) error
	CanDeleteCardHolder(cardHolder Operant) error

	CanUpdateDoor(door Operant, field string, value interface{}) error
	CanAddDoor(door Operant) error
	CanDeleteDoor(controller Operant) error
}

type Operant interface {
	AsRuleEntity() interface{}
}

func NewAuthProvider(config string, loginExpiry, sessionExpiry string) (IAuth, error) {
	return NewLocalAuthProvider(config, loginExpiry, sessionExpiry)
}
