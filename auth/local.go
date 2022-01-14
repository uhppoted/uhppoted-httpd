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

const KEY_LENGTH = 32 // 256 bits
const SALT_LENGTH = 32

type Local struct {
	key       []byte
	users     map[string]*user
	resources []resource

	loginExpiry   time.Duration
	sessionExpiry time.Duration
	file          string

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

func NewLocalAuthProvider(file string, loginExpiry, sessionExpiry string) (*Local, error) {
	provider := Local{
		key:  make([]byte, KEY_LENGTH),
		file: file,
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

	if _, err := io.ReadFull(rand.Reader, provider.key); err != nil {
		return nil, err
	}

	provider.watch(file)

	fmt.Printf(">>>>>>>>>>>>>>>> KEY: %X\n", provider.key)

	return &provider, nil
}

func (p *Local) Preauthenticate(loginId uuid.UUID) (string, error) {
	secret := p.copyKey()
	expiry := p.loginExpiry

	fmt.Printf(">>>>>>>>>>>>>>>> SECRET/LOGIN: %X\n", secret)

	signer, err := jwt.NewSignerHS(jwt.HS256, secret)
	if err != nil {
		return "", err
	}

	UUID, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}

	salt := make([]byte, SALT_LENGTH)
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

	return token.String(), nil
}

func (p *Local) Authorize(uid, pwd string, sessionId uuid.UUID) (string, error) {
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

func (p *Local) Verify(tokenType TokenType, cookie string) (*uuid.UUID, error) {
	secret := p.copyKey()

	verifier, err := jwt.NewVerifierHS(jwt.HS256, secret)
	if err != nil {
		return nil, err
	}

	token, err := jwt.ParseAndVerifyString(cookie, verifier)
	if err != nil {
		return nil, err
	}

	if err := verifier.Verify(token.Payload(), token.Signature()); err != nil {
		return nil, err
	}

	var claims claims
	if err := json.Unmarshal(token.RawClaims(), &claims); err != nil {
		return nil, err
	}

	if !claims.IsValidAt(time.Now()) {
		return nil, fmt.Errorf("JWT token expired")
	}

	switch tokenType {
	case Login:
		if !claims.IsForAudience("login") {
			return nil, fmt.Errorf("Invalid audience in JWT claims")
		} else if claims.Login == nil {
			return nil, fmt.Errorf("Invalid login token")
		} else {
			return &claims.Login.LoginId, nil
		}

	case Session:
		if !claims.IsForAudience("admin") {
			return nil, fmt.Errorf("Invalid audience in JWT claims")
		} else if claims.Session == nil {
			return nil, fmt.Errorf("Invalid session token")
		} else {
			return &claims.Session.SessionId, nil
		}
	}

	return nil, nil
}

func (p *Local) Authenticated(cookie string) (string, string, *uuid.UUID, error) {
	secret := p.copyKey()

	verifier, err := jwt.NewVerifierHS(jwt.HS256, secret)
	if err != nil {
		return "", "", nil, err
	}

	token, err := jwt.ParseAndVerifyString(cookie, verifier)
	if err != nil {
		return "", "", nil, err
	}

	if err := verifier.Verify(token.Payload(), token.Signature()); err != nil {
		return "", "", nil, err
	}

	var claims claims
	if err := json.Unmarshal(token.RawClaims(), &claims); err != nil {
		return "", "", nil, err
	}

	if !claims.IsForAudience("admin") {
		return "", "", nil, fmt.Errorf("Invalid audience in JWT claims")
	}

	if !claims.IsValidAt(time.Now()) {
		return "", "", nil, fmt.Errorf("JWT token expired")
	}

	if claims.Session == nil {
		return "", "", nil, fmt.Errorf("Invalid session token")
	}

	return claims.Session.LoggedInAs, claims.Session.Role, &claims.Session.SessionId, nil
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

func (p *Local) copyKey() []byte {
	p.guard.Lock()
	defer p.guard.Unlock()

	k := make([]byte, KEY_LENGTH)

	copy(k, p.key)

	return k
}
