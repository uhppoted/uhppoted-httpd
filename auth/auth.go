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
	Authorized(token, resource string) (string, error)
	GetLoginId(token string) (*uuid.UUID, error)
	GetSessionId(token string) (*uuid.UUID, error)
}

func NewAuthProvider(config string, loginExpiry, sessionExpiry string) (IAuth, error) {
	return NewLocalAuthProvider(config, loginExpiry, sessionExpiry)
}
