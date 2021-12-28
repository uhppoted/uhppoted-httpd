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
	Validate(uid, pwd string) error
	Verify(tokenType TokenType, token string) error
	Authenticated(token string) (string, string, error)
	AuthorisedX(uid, role, resource string) error
	Authorized(token, resource string) (string, string, error)
	GetLoginId(token string) (*uuid.UUID, error)
	GetSessionId(token string) (*uuid.UUID, error)

	Store(uid, pwd, role string) error
	Save() error
}

type OpAuth interface {
	UID() string

	CanView(ruleset string, o Operant, field string, value interface{}) error

	CanUpdateInterface(iface Operant, field string, value interface{}) error

	CanAddController(controller Operant) error
	CanUpdateController(controller Operant, field string, value interface{}) error
	CanDeleteController(controller Operant) error

	CanAddCard(card Operant) error
	CanUpdateCard(card Operant, field string, value interface{}) error
	CanDeleteCard(card Operant) error

	CanUpdateDoor(door Operant, field string, value interface{}) error
	CanAddDoor(door Operant) error
	CanDeleteDoor(door Operant) error

	CanAddGroup(group Operant) error
	CanUpdateGroup(group Operant, field string, value interface{}) error
	CanDeleteGroup(group Operant) error
}

type Operant interface {
	AsRuleEntity() interface{}
}

func NewAuthProvider(config string, loginExpiry, sessionExpiry string) (IAuth, error) {
	return NewLocalAuthProvider(config, loginExpiry, sessionExpiry)
}
