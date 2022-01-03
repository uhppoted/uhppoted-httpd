package auth

import (
	"github.com/google/uuid"
)

type TokenType int

const (
	Login   TokenType = iota // 0
	Session                  // 1
)

type RuleSet int

const (
	Interfaces = iota
	Controllers
	Doors
	Cards
	Groups
	Events
	Logs
)

func (r RuleSet) String() string {
	return [...]string{"interfaces", "controllers", "doors", "cards", "groups", "events", "logs"}[r]
}

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

	CanView(r RuleSet, o Operant, field string, value interface{}) error
	CanAdd(r RuleSet, o Operant) error
	CanUpdate(r RuleSet, o Operant, field string, value interface{}) error
	CanDelete(r RuleSet, o Operant) error
}

type Operant interface {
	AsRuleEntity() (string, interface{})
}

func NewAuthProvider(config string, loginExpiry, sessionExpiry string) (IAuth, error) {
	return NewLocalAuthProvider(config, loginExpiry, sessionExpiry)
}
