package auth

type IAuth interface {
	Authorize(uid, pwd string) (string, error)
	Verify(token string) error
	Authorized(token, resource string) error
	User(token string) (string, error)
}

func NewAuthProvider(config string, sessionExpiry string) (IAuth, error) {
	return NewLocalAuthProvider(config, sessionExpiry)
}
