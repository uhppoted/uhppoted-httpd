package auth

import (
	"github.com/google/uuid"
)

type IAuth interface {
	Authorize(uid, pwd string, sessionId uuid.UUID) (string, error)
	Verify(token string) error
	Authorized(token, resource string) error
	Session(token string) (*uuid.UUID, error)
}

func NewAuthProvider(config string, sessionExpiry string) (IAuth, error) {
	return NewLocalAuthProvider(config, sessionExpiry)
}
