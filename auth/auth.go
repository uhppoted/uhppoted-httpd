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
	Options(uid, role string) Options
	AdminRole() string
}

type IUser interface {
	Password() ([]byte, string)
	OTPKey() string
	Role() string
	Locked() bool
	IsDeleted() bool
}

type Authorizator struct {
	uid  string
	role string
	OpAuth
}

type OpAuth interface {
	CanView(o Operant, field string, value any, rulesets ...RuleSet) error
	CanAdd(o Operant, rulesets ...RuleSet) error
	CanUpdate(o Operant, field string, value any, rulesets ...RuleSet) error
	CanDelete(o Operant, rulesets ...RuleSet) error
	CanCache(o Operant, field string, cache string, rulesets ...RuleSet) error
}

type Options struct {
	OTP struct {
		Allowed bool
		Enabled bool
	}
}

type Operant interface {
	AsRuleEntity() (string, any)
	CacheKey() string
}

func NewAuthorizator(uid, role string) *Authorizator {
	return &Authorizator{
		uid:  uid,
		role: role,
		OpAuth: &authorizator{
			uid:  uid,
			role: role,
		},
	}
}

func UID(a *Authorizator) string {
	if a != nil {
		return a.uid
	}

	return ""
}
