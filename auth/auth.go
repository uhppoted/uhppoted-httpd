package auth

import (
	"reflect"
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

type Authorizator struct {
	uid  string
	role string
	OpAuth
}

type OpAuth interface {
	CanView(o Operant, field string, value interface{}, rulesets ...RuleSet) error
	CanAdd(o Operant, rulesets ...RuleSet) error
	CanUpdate(o Operant, field string, value interface{}, rulesets ...RuleSet) error
	CanDelete(o Operant, rulesets ...RuleSet) error
}

func IsNil(v any) bool {
	if v == nil {
		return true
	}

	switch reflect.TypeOf(v).Kind() {
	case reflect.Ptr,
		reflect.Map,
		reflect.Array,
		reflect.Chan,
		reflect.Slice:
		return reflect.ValueOf(v).IsNil()
	}

	return false
}

type Operant interface {
	AsRuleEntity() (string, interface{})
}

func NewAuthorizator(uid, role string) *Authorizator {
	return &Authorizator{
		uid:    uid,
		role:   role,
		OpAuth: &authorizator{},
	}
}

func UID(a *Authorizator) string {
	if a != nil {
		return a.uid
	}

	return ""
}
