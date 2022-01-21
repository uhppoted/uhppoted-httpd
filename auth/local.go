package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/cristalhq/jwt/v3"
	"github.com/google/uuid"
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
	IDLETIME:    1 * time.Minute,  // Mark untouched sessions and logins as idle after 10 minutes
	SWEEP:       60 * time.Second, // Sweep session and login caches every minute
}

type Local struct {
	keys      [][]byte
	users     map[string]*user
	resources []resource

	loginExpiry   time.Duration
	sessionExpiry time.Duration
	file          string

	logins   map[uuid.UUID]time.Time
	sessions map[uuid.UUID]time.Time

	guard sync.Mutex
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

type session struct {
	LoggedInAs string    `json:"uid,omitempty"`
	SessionId  uuid.UUID `json:"session.id,omitempty"`
	Role       string    `json:"session.role,omitempty"`
}

type user struct {
	Salt     salt   `json:"salt"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type resource struct {
	Path       *regexp.Regexp `json:"path"`
	Authorised *regexp.Regexp `json:"authorised"`
}

func NewLocalAuthProvider(file string, loginExpiry, sessionExpiry string) (*Local, error) {
	provider := Local{
		keys:          make([][]byte, constants.KEYS),
		loginExpiry:   1 * time.Minute,
		sessionExpiry: 60 * time.Minute,
		file:          file,
		logins:        map[uuid.UUID]time.Time{},
		sessions:      map[uuid.UUID]time.Time{},
	}

	if err := provider.load(file); err != nil {
		return nil, err
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

	provider.watch(file)

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
	secret := p.copyKey()
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

	p.logins[loginId] = time.Now()

	return token.String(), nil
}

func (p *Local) Authenticate(uid, pwd string) (string, error) {
	p.guard.Lock()
	users := p.users
	expiry := p.sessionExpiry
	p.guard.Unlock()

	secret := p.copyKey()

	u, ok := users[uid]
	if !ok {
		return "", fmt.Errorf("Invalid login credentials")
	}

	h := sha256.New()
	h.Write(u.Salt)
	h.Write([]byte(pwd))

	hash := fmt.Sprintf("%0x", h.Sum(nil))
	if hash != u.Password {
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
			Role:       u.Role,
		},
	}

	token, err := jwt.NewBuilder(signer).Build(claims)
	if err != nil {
		return "", err
	}

	p.sessions[sessionId] = time.Now()

	return token.String(), nil
}

func (p *Local) Validate(uid, pwd string) error {
	p.guard.Lock()
	defer p.guard.Unlock()

	users := p.users
	u, ok := users[uid]
	if !ok {
		return fmt.Errorf("invalid user ID or password")
	}

	h := sha256.New()
	h.Write(u.Salt)
	h.Write([]byte(pwd))

	hash := fmt.Sprintf("%0x", h.Sum(nil))
	if hash != u.Password {
		return fmt.Errorf("invalid user ID or password")
	}

	return nil
}

func (p *Local) Invalidate(tokenType TokenType, cookie string) error {
	//	token, _, err := p.getToken(cookie)
	//	if err != nil {
	//		return err
	//	}
	//
	//	var claims claims
	//	if err := json.Unmarshal(token.RawClaims(), &claims); err != nil {
	//		return err
	//	}
	//
	//	switch tokenType {
	//	case Login:
	//		delete(p.logins, claims.Login.LoginId)
	//
	//	case Session:
	//		delete(p.sessions, claims.Session.SessionId)
	//	}

	return nil
}

func (p *Local) Store(uid, pwd, role string) error {
	if strings.TrimSpace(uid) == "" {
		return fmt.Errorf("Invalid user ID or password")
	}

	p.guard.Lock()
	defer p.guard.Unlock()

	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return err
	}

	k := strings.TrimSpace(uid)
	h := sha256.New()
	h.Write(salt)
	h.Write([]byte(pwd))

	hash := fmt.Sprintf("%0x", h.Sum(nil))

	p.users[k] = &user{
		Salt:     salt,
		Password: hash,
		Role:     role,
	}

	return nil
}

func (p *Local) Verify(tokenType TokenType, cookie string) error {
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
	case Login:
		if !claims.IsForAudience("login") {
			return fmt.Errorf("Invalid audience in JWT claims")
		} else if claims.Login == nil {
			return fmt.Errorf("Invalid login token")
		} else if _, ok := p.logins[claims.Login.LoginId]; !ok {
			return fmt.Errorf("No extant login token for %v", claims.Login.LoginId)
		} else if err := p.extant(Login, claims.Login.LoginId); err != nil {
			return err
		} else {
			return nil
		}

	case Session:
		if !claims.IsForAudience("admin") {
			return fmt.Errorf("Invalid audience in JWT claims")
		} else if claims.Session == nil {
			return fmt.Errorf("Invalid session token")
		} else if err := p.extant(Session, claims.Session.SessionId); err != nil {
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

	if err := p.extant(Session, claims.Session.SessionId); err != nil {
		return "", "", "", err
	}

	p.sessions[claims.Session.SessionId] = time.Now()

	if keyID == 1 {
		return claims.Session.LoggedInAs, claims.Session.Role, "", nil
	}

	secret := p.copyKey()

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

func (p *Local) Authorised(uid, role, resource string) error {
	for _, r := range p.resources {
		if r.Path.Match([]byte(resource)) && r.Authorised.Match([]byte(role)) {
			return nil
		}
	}

	return fmt.Errorf("%v not authorized for %s", uid, resource)
}

func (p *Local) load(file string) error {
	bytes, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	return p.deserialize(bytes)
}

func (p *Local) Save() error {
	b, err := p.serialize()
	if err != nil {
		return err
	}

	tmp, err := os.CreateTemp("", "uhppoted-auth.*")
	if err != nil {
		return err
	}

	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(b); err != nil {
		return err
	}

	if err := tmp.Close(); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(p.file), 0770); err != nil {
		return err
	}

	return os.Rename(tmp.Name(), p.file)
}

func (p *Local) serialize() ([]byte, error) {
	p.guard.Lock()
	defer p.guard.Unlock()

	serializable := struct {
		Users     map[string]*user `json:"users"`
		Resources []resource       `json:"resources"`
	}{
		Users:     p.users,
		Resources: p.resources,
	}

	return json.MarshalIndent(serializable, "", "  ")
}

func (p *Local) deserialize(bytes []byte) error {
	serializable := struct {
		Users     map[string]*user `json:"users"`
		Resources []resource       `json:"resources"`
	}{
		Users:     map[string]*user{},
		Resources: []resource{},
	}

	p.guard.Lock()
	defer p.guard.Unlock()
	if err := json.Unmarshal(bytes, &serializable); err != nil {
		return err
	}

	p.users = serializable.Users
	p.resources = serializable.Resources

	return nil
}

// NOTE: interim file watcher implementation pending fsnotify in Go v?.?
//       (https://github.com/fsnotify/fsnotify requires workarounds for
//        files updated atomically by renaming)
func (p *Local) watch(filepath string) {
	go func() {
		finfo, err := os.Stat(filepath)
		if err != nil {
			log.Printf("ERROR Failed to get file information for '%s': %v", filepath, err)
			return
		}

		lastModified := finfo.ModTime()
		logged := false
		for {
			time.Sleep(2500 * time.Millisecond)
			finfo, err := os.Stat(filepath)
			if err != nil {
				if !logged {
					log.Printf("ERROR Failed to get file information for '%s': %v", filepath, err)
					logged = true
				}

				continue
			}

			logged = false
			if finfo.ModTime() != lastModified {
				log.Printf("INFO  Reloading information from %s\n", filepath)

				err := p.load(filepath)
				if err != nil {
					log.Printf("ERROR Failed to reload information from %s: %v", filepath, err)
					continue
				}

				log.Printf("INFO  Updated auth DB from %s", filepath)
				lastModified = finfo.ModTime()
			}
		}
	}()
}

func (p *Local) getToken(cookie string) (*jwt.Token, int, error) {
	p.guard.Lock()
	defer p.guard.Unlock()

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

func (p *Local) copyKey() []byte {
	p.guard.Lock()
	defer p.guard.Unlock()

	key := make([]byte, constants.KEY_LENGTH)

	copy(key, p.keys[0])

	return key
}

func (p *Local) regenerate() {
	key, err := genKey()
	if err != nil {
		log.Printf("%-5v Failed to regenerate session secret (%v)", "ERROR", err)
		return
	}

	p.guard.Lock()
	defer p.guard.Unlock()

	for i := 1; i < len(p.keys); i++ {
		p.keys[i] = p.keys[i-1]
	}

	p.keys[0] = key

	log.Printf("%-5v Regenerated session secret", "INFO")
}

func (p *Local) extant(tokenType TokenType, id uuid.UUID) error {
	p.guard.Lock()
	defer p.guard.Unlock()

	cutoff := time.Now().Add(-constants.IDLETIME)

	switch tokenType {
	case Login:

		if touched, ok := p.logins[id]; !ok {
			return fmt.Errorf("No extant login for login ID '%v'", id)
		} else if touched.Before(cutoff) {
			return fmt.Errorf("Login '%v' expired", id)
		}

	case Session:
		if touched, ok := p.sessions[id]; !ok {
			return fmt.Errorf("No extant session for session ID '%v'", id)
		} else if touched.Before(cutoff) {
			return fmt.Errorf("Session '%v' expired", id)
		}
	}

	return nil
}

func (p *Local) sweep() {
	p.guard.Lock()
	defer p.guard.Unlock()

	caches := []struct {
		cache  map[uuid.UUID]time.Time
		format string
	}{
		{p.logins, "%-5v Deleted idle login %v"},
		{p.sessions, "%-5v Deleted idle session %v"},
	}

	cutoff := time.Now().Add(-2 * constants.IDLETIME)

	for _, c := range caches {
		list := []uuid.UUID{}
		for k, touched := range c.cache {
			if touched.Before(cutoff) {
				list = append(list, k)
			}
		}

		for _, k := range list {
			delete(c.cache, k)
			log.Printf(c.format, "INFO", k)
		}
	}
}

func genKey() ([]byte, error) {
	key := make([]byte, constants.KEY_LENGTH)

	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}

	return key, nil
}

func (s *salt) MarshalJSON() ([]byte, error) {
	bytes := []byte{}

	if s != nil {
		bytes = []byte(*s)
	}

	return json.Marshal(hex.EncodeToString(bytes[:]))
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

func (r resource) MarshalJSON() ([]byte, error) {
	object := struct {
		Path       string `json:"path"`
		Authorised string `json:"authorised"`
	}{
		Path:       fmt.Sprintf("%v", r.Path),
		Authorised: fmt.Sprintf("%v", r.Authorised),
	}

	return json.Marshal(object)
}

func (r *resource) UnmarshalJSON(bytes []byte) error {
	x := struct {
		Path       string `json:"path"`
		Authorised string `json:"authorised"`
	}{}

	err := json.Unmarshal(bytes, &x)
	if err != nil {
		return err
	}

	path, err := regexp.Compile(fmt.Sprintf("^%v$", x.Path))
	if err != nil {
		return err
	}

	authorised, err := regexp.Compile(x.Authorised)
	if err != nil {
		return err
	}

	r.Path = path
	r.Authorised = authorised

	return nil
}
