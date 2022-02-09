package auth

import ()

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
	Users
)

func (r RuleSet) String() string {
	return [...]string{"interfaces", "controllers", "doors", "cards", "groups", "events", "logs"}[r]
}

type IAuth interface {
	Preauthenticate() (string, error)
	Authenticate(uid, pwd string) (string, error)
	Validate(uid, pwd string) error
	Verify(tokenType TokenType, token string) error
	Authenticated(token string) (string, string, string, error)
	Authorised(uid, role, resource string) error
	Invalidate(tokenType TokenType, token string) error

	Store(uid, pwd, role string) error
	Save() error
}

type OpAuth interface {
	UID() string

	CanView(o Operant, field string, value interface{}, rulesets ...RuleSet) error
	CanAdd(o Operant, rulesets ...RuleSet) error
	CanUpdate(o Operant, field string, value interface{}, rulesets ...RuleSet) error
	CanDelete(o Operant, rulesets ...RuleSet) error
}

type Operant interface {
	AsRuleEntity() (string, interface{})
}

func NewAuthProvider(config string, loginExpiry, sessionExpiry string) (IAuth, error) {
	return NewLocalAuthProvider(config, loginExpiry, sessionExpiry)
}
