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
	return [...]string{"interfaces", "controllers", "doors", "cards", "groups", "events", "logs", "users"}[r]
}

type IAuthenticate interface {
	Preauthenticate() (string, error)
	Authenticate(uid, pwd string) (string, error)
	Validate(uid, pwd string) error
	Verify(tokenType TokenType, token string) error
	Authenticated(token string) (string, string, string, error)
	Invalidate(tokenType TokenType, token string) error
}

type IUser interface {
	Password() ([]byte, string)
	Role() string
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
