package local

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/cristalhq/jwt/v3"
	"github.com/google/uuid"

	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/auth/otp"
	"github.com/uhppoted/uhppoted-httpd/system"
)

var constants = struct {
	KEYS        int           // Number of historical session secrets
	KEY_LENGTH  int           // bytes
	SALT_LENGTH int           // bytes
	REGENERATE  time.Duration // Interval at which the internal secret keys are regenerated
	IDLETIME    time.Duration // Interval after which an untouched session is marked 'idle'
	SWEEP       time.Duration // Interval at which the session list is 'swept'

}{
	KEYS:        2,                // 2 historical session secrets
	KEY_LENGTH:  256 / 8,          // 256 bits
	SALT_LENGTH: 256 / 8,          // 256 bits
	REGENERATE:  15 * time.Minute, // Regenerate secret keys at 15 minute intervals
	IDLETIME:    30 * time.Minute, // FIXME (disabled for OTP)  Mark untouched sessions and logins as idle after 10 minutes
	SWEEP:       60 * time.Second, // Sweep session and login caches every minute
}

type Local struct {
	keys          [][]byte
	users         map[string]*user
	loginExpiry   time.Duration
	sessionExpiry time.Duration
	allowOTPLogin bool

	logins   sessions
	sessions sessions

	sync.RWMutex
}

type sessions struct {
	list map[uuid.UUID]time.Time
	sync.Mutex
}

type salt []byte

type claims struct {
	jwt.StandardClaims
	Login   *login   `json:"login,omitempty"`
	Session *session `json:"session,omitempty"`
}

type login struct {
	LoginId uuid.UUID `json:"login.id,omitempty"`
	Salt    []byte    `json:"login.salt,omitempty"`
}

type user struct {
	Salt     salt   `json:"salt"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type session struct {
	LoggedInAs string    `json:"uid,omitempty"`
	SessionId  uuid.UUID `json:"session.id,omitempty"`
	Role       string    `json:"session.role,omitempty"`
}

func NewAuthProvider(file string, loginExpiry, sessionExpiry string, allowOTPLogin bool) (*Local, error) {
	provider := Local{
		keys:  make([][]byte, constants.KEYS),
		users: map[string]*user{},

		logins: sessions{
			list: map[uuid.UUID]time.Time{},
		},

		sessions: sessions{
			list: map[uuid.UUID]time.Time{},
		},

		sessionExpiry: 60 * time.Minute,
		loginExpiry:   1 * time.Minute,
		allowOTPLogin: allowOTPLogin,
	}

	if f, err := os.Open(file); err != nil {
		return nil, err
	} else {
		defer f.Close()
		if err := provider.deserialize(f); err != nil {
			return nil, err
		}
	}

	if t, err := time.ParseDuration(loginExpiry); err != nil {
		return nil, err
	} else {
		provider.loginExpiry = t
	}

	if t, err := time.ParseDuration(sessionExpiry); err != nil {
		return nil, err
	} else {
		provider.sessionExpiry = t
	}

	if key, err := genKey(); err != nil {
		return nil, err
	} else {
		provider.keys[0] = key
	}

	regen := time.Tick(constants.REGENERATE)
	sweep := time.Tick(constants.SWEEP)
	go func() {
		for {
			select {
			case <-regen:
				go func() {
					provider.regenerate()
				}()

			case <-sweep:
				go func() {
					provider.sweep()
				}()
			}
		}
	}()

	return &provider, nil
}

func (p *Local) Preauthenticate() (string, error) {
	p.RLock()
	defer p.RUnlock()

	secret := p.keys[0]
	expiry := p.loginExpiry

	signer, err := jwt.NewSignerHS(jwt.HS256, secret)
	if err != nil {
		return "", err
	}

	UUID, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}

	loginId, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}

	salt := make([]byte, constants.SALT_LENGTH)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", err
	}

	claims := &claims{
		StandardClaims: jwt.StandardClaims{
			ID:        UUID.String(),
			Audience:  []string{"login"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expiry))),
		},
		Login: &login{
			LoginId: loginId,
			Salt:    salt,
		},
	}

	token, err := jwt.NewBuilder(signer).Build(claims)
	if err != nil {
		return "", err
	}

	p.touched(auth.Login, loginId)

	return token.String(), nil
}

func (p *Local) Authenticate(uid, pwd string) (string, error) {
	p.RLock()
	defer p.RUnlock()

	secret := p.keys[0]
	expiry := p.sessionExpiry

	var salt []byte
	var password string
	var role string

	if u, ok := system.User(uid); ok && u != nil {
		salt, password = u.Password()
		role = u.Role()
	} else {
		if u, ok := p.users[uid]; !ok {
			return "", fmt.Errorf("Invalid login credentials")
		} else {
			salt = u.Salt
			password = u.Password
			role = u.Role
		}
	}

	h := sha256.New()
	h.Write(salt)
	h.Write([]byte(pwd))
	hash := fmt.Sprintf("%0x", h.Sum(nil))

	if hash != password && (!p.allowOTPLogin || !otp.Verify(uid, pwd)) {
		return "", fmt.Errorf("Invalid login credentials")
	}

	signer, err := jwt.NewSignerHS(jwt.HS256, secret)
	if err != nil {
		return "", err
	}

	UUID, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}

	sessionId, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}

	claims := &claims{
		StandardClaims: jwt.StandardClaims{
			ID:        UUID.String(),
			Audience:  []string{"admin"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expiry))),
		},
		Session: &session{
			LoggedInAs: uid,
			SessionId:  sessionId,
			Role:       role,
		},
	}

	token, err := jwt.NewBuilder(signer).Build(claims)
	if err != nil {
		return "", err
	}

	p.touched(auth.Session, sessionId)

	return token.String(), nil
}

func (p *Local) Validate(uid, pwd string) error {
	var salt []byte
	var password string

	if u, ok := system.User(uid); ok && u != nil {
		salt, password = u.Password()
	} else {
		if u, ok := p.users[uid]; !ok {
			return fmt.Errorf("invalid user ID or password")
		} else {
			salt = u.Salt
			password = u.Password
		}
	}

	h := sha256.New()
	h.Write(salt)
	h.Write([]byte(pwd))

	hash := fmt.Sprintf("%0x", h.Sum(nil))
	if hash != password {
		return fmt.Errorf("invalid user ID or password")
	}

	return nil
}

func (p *Local) Invalidate(tokenType auth.TokenType, cookie string) error {
	token, _, err := p.getToken(cookie)
	if err != nil {
		return err
	}

	var claims claims
	if err := json.Unmarshal(token.RawClaims(), &claims); err != nil {
		return err
	}

	switch tokenType {
	case auth.Login:
		p.logins.delete(claims.Login.LoginId)

	case auth.Session:
		p.sessions.delete(claims.Session.SessionId)
	}

	return nil
}

func (p *Local) Verify(tokenType auth.TokenType, cookie string) error {
	token, _, err := p.getToken(cookie)
	if err != nil {
		return err
	}

	var claims claims
	if err := json.Unmarshal(token.RawClaims(), &claims); err != nil {
		return err
	}

	if !claims.IsValidAt(time.Now()) {
		return fmt.Errorf("JWT token expired")
	}

	switch tokenType {
	case auth.Login:
		if !claims.IsForAudience("login") {
			return fmt.Errorf("Invalid audience in JWT claims")
		} else if claims.Login == nil {
			return fmt.Errorf("Invalid login token")
		} else if err := p.extant(auth.Login, claims.Login.LoginId); err != nil {
			return err
		} else {
			return nil
		}

	case auth.Session:
		if !claims.IsForAudience("admin") {
			return fmt.Errorf("Invalid audience in JWT claims")
		} else if claims.Session == nil {
			return fmt.Errorf("Invalid session token")
		} else if err := p.extant(auth.Session, claims.Session.SessionId); err != nil {
			return err
		} else {
			return nil
		}
	}

	return nil
}

func (p *Local) Authenticated(cookie string) (string, string, string, error) {
	token, keyID, err := p.getToken(cookie)
	if err != nil {
		return "", "", "", err
	}

	var claims claims
	if err := json.Unmarshal(token.RawClaims(), &claims); err != nil {
		return "", "", "", err
	}

	if !claims.IsForAudience("admin") {
		return "", "", "", fmt.Errorf("Invalid audience in JWT claims")
	}

	if !claims.IsValidAt(time.Now()) {
		return "", "", "", fmt.Errorf("JWT token expired")
	}

	if claims.Session == nil {
		return "", "", "", fmt.Errorf("Invalid session token")
	}

	if err := p.extant(auth.Session, claims.Session.SessionId); err != nil {
		return "", "", "", err
	}

	p.touched(auth.Session, claims.Session.SessionId)

	if keyID == 1 {
		return claims.Session.LoggedInAs, claims.Session.Role, "", nil
	}

	p.RLock()
	defer p.RUnlock()
	secret := p.keys[0]

	signer, err := jwt.NewSignerHS(jwt.HS256, secret)
	if err != nil {
		return "", "", "", err
	}

	token2, err := jwt.NewBuilder(signer).Build(claims)
	if err != nil {
		return "", "", "", err
	}

	return claims.Session.LoggedInAs, claims.Session.Role, token2.String(), nil
}

func (p *Local) Options(uid string) auth.Options {
	return auth.Options{
		OTP: struct {
			Allowed bool
			Enabled bool
		}{
			Allowed: p.allowOTPLogin,
			Enabled: otp.Enabled(uid),
		},
	}
}

func (p *Local) deserialize(r io.Reader) error {
	bytes, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	serializable := struct {
		Users map[string]*user `json:"users"`
	}{
		Users: map[string]*user{},
	}

	if err := json.Unmarshal(bytes, &serializable); err != nil {
		return err
	}

	p.users = serializable.Users

	return nil
}

func (p *Local) getToken(cookie string) (*jwt.Token, int, error) {
	p.RLock()
	defer p.RUnlock()

	secrets := p.keys
	for ix, secret := range secrets {
		// NOTE: jwt.NewVerifier returns an error if the secret is nil so this is just a courtesy
		//       thing to avoid a "jwt: key is nil" warning in the log when the HTTPD server has
		//       been restarted and the browser does a refresh with a no longer valid session
		//       cookie.
		if secret == nil {
			continue
		}

		verifier, err := jwt.NewVerifierHS(jwt.HS256, secret)
		if err != nil {
			return nil, 0, err
		}

		token, err := jwt.ParseAndVerifyString(cookie, verifier)
		if err != nil {
			continue
		}

		if err := verifier.Verify(token.Payload(), token.Signature()); err != nil {
			return nil, 0, err
		}

		return token, ix + 1, nil
	}

	return nil, 0, fmt.Errorf("JWT signature is not valid")
}

func (p *Local) regenerate() {
	p.Lock()
	defer p.Unlock()

	key, err := genKey()
	if err != nil {
		log.Printf("%-5v Failed to regenerate session secret (%v)", "ERROR", err)
		return
	}

	for i := 1; i < len(p.keys); i++ {
		p.keys[i] = p.keys[i-1]
	}

	p.keys[0] = key

	log.Printf("%-5v Regenerated session secret", "INFO")
}

func (p *Local) extant(tokenType auth.TokenType, id uuid.UUID) error {
	switch tokenType {
	case auth.Login:
		return p.logins.extant(id)

	case auth.Session:
		return p.sessions.extant(id)
	}

	return nil
}

func (p *Local) touched(tt auth.TokenType, uuid uuid.UUID) {
	switch tt {
	case auth.Login:
		p.logins.touched(uuid)
	case auth.Session:
		p.sessions.touched(uuid)
	}
}

func (p *Local) sweep() {
	caches := []struct {
		cache  *sessions
		format string
	}{
		{&p.logins, "%-5v Deleted idle login %v"},
		{&p.sessions, "%-5v Deleted idle session %v"},
	}

	cutoff := time.Now().Add(-2 * constants.IDLETIME)

	for _, c := range caches {
		c.cache.Lock()

		list := []uuid.UUID{}
		for k, touched := range c.cache.list {
			if touched.Before(cutoff) {
				list = append(list, k)
			}
		}

		for _, k := range list {
			delete(c.cache.list, k)
			log.Printf(c.format, "INFO", k)
		}

		c.cache.Unlock()
	}
}

func genKey() ([]byte, error) {
	key := make([]byte, constants.KEY_LENGTH)

	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}

	return key, nil
}

func (ss *sessions) touched(uuid uuid.UUID) {
	ss.Lock()
	defer ss.Unlock()

	ss.list[uuid] = time.Now()
}

func (ss *sessions) extant(uuid uuid.UUID) error {
	cutoff := time.Now().Add(-constants.IDLETIME)

	ss.Lock()
	defer ss.Unlock()

	if touched, ok := ss.list[uuid]; !ok {
		return fmt.Errorf("No extant session for ID '%v'", uuid)
	} else if touched.Before(cutoff) {
		return fmt.Errorf("Session '%v' expired", uuid)
	}

	return nil
}

func (ss *sessions) delete(uuid uuid.UUID) {
	ss.Lock()
	defer ss.Unlock()

	delete(ss.list, uuid)
}

func (s *salt) UnmarshalJSON(bytes []byte) error {
	re := regexp.MustCompile(`^"([0-9a-fA-F]*)"$`)
	match := re.FindSubmatch(bytes)

	if len(match) < 2 {
		return fmt.Errorf("Invalid salt '%s'", string(bytes))
	}

	b, err := hex.DecodeString(string(match[1]))
	if err != nil {
		return err
	}

	*s = b

	return nil
}
