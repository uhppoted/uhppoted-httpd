package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/cristalhq/jwt/v3"
	"github.com/google/uuid"
)

const SALT_LENGTH = 32

type Local struct {
	Key           string           `json:"key"`
	Users         map[string]*user `json:"users"`
	Resources     []resource       `json:"resources"`
	loginExpiry   time.Duration
	sessionExpiry time.Duration

	guard    sync.Mutex
	resource []resource
}

type salt []byte

type login struct {
	jwt.StandardClaims
	LoggedInAs string    `json:"uid"`
	LoginId    uuid.UUID `json:"login-id"`
	Salt       []byte    `json:"salt"`
}

type session struct {
	jwt.StandardClaims
	LoggedInAs string    `json:"uid"`
	SessionId  uuid.UUID `json:"session-id"`
	Role       string    `json:"role"`
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

func (r *resource) UnmarshalJSON(bytes []byte) error {
	x := struct {
		Path       string `json:"path"`
		Authorised string `json:"authorised"`
	}{}

	err := json.Unmarshal(bytes, &x)
	if err != nil {
		return err
	}

	path, err := regexp.Compile(x.Path)
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
	provider := Local{}
	if err := provider.load(file); err != nil {
		return nil, err
	}

	{
		t, err := time.ParseDuration(loginExpiry)
		if err != nil {
			return nil, err
		}

		provider.loginExpiry = t
	}

	{
		t, err := time.ParseDuration(sessionExpiry)
		if err != nil {
			return nil, err
		}

		provider.sessionExpiry = t
	}

	provider.watch(file)

	return &provider, nil
}

func (p *Local) Preauthenticate(loginId uuid.UUID) (string, error) {
	p.guard.Lock()
	secret := []byte(p.Key)
	expiry := p.loginExpiry
	p.guard.Unlock()

	signer, err := jwt.NewSignerHS(jwt.HS256, secret)
	if err != nil {
		return "", err
	}

	uuid, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}

	salt := make([]byte, SALT_LENGTH)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", err
	}

	claims := &login{
		StandardClaims: jwt.StandardClaims{
			ID:        uuid.String(),
			Audience:  []string{"login"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expiry))),
		},
		LoginId: loginId,
		Salt:    salt,
	}

	token, err := jwt.NewBuilder(signer).Build(claims)
	if err != nil {
		return "", err
	}

	return token.String(), nil
}

func (p *Local) Authorize(uid, pwd string, sessionId uuid.UUID) (string, error) {
	p.guard.Lock()
	secret := []byte(p.Key)
	users := p.Users
	expiry := p.sessionExpiry
	p.guard.Unlock()

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

	uuid, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}

	claims := &session{
		StandardClaims: jwt.StandardClaims{
			ID:        uuid.String(),
			Audience:  []string{"admin"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expiry))),
		},
		LoggedInAs: uid,
		SessionId:  sessionId,
		Role:       u.Role,
	}

	token, err := jwt.NewBuilder(signer).Build(claims)
	if err != nil {
		return "", err
	}

	return token.String(), nil
}

func (p *Local) Verify(tokenType TokenType, cookie string) error {
	p.guard.Lock()
	secret := []byte(p.Key)
	p.guard.Unlock()

	verifier, err := jwt.NewVerifierHS(jwt.HS256, secret)
	if err != nil {
		return err
	}

	token, err := jwt.ParseAndVerifyString(cookie, verifier)
	if err != nil {
		return err
	}

	if err := verifier.Verify(token.Payload(), token.Signature()); err != nil {
		return err
	}

	var claims session
	if err := json.Unmarshal(token.RawClaims(), &claims); err != nil {
		return err
	}

	switch tokenType {
	case Login:
		if !claims.IsForAudience("login") {
			return fmt.Errorf("Invalid audience in JWT claims")
		}

	case Session:
		if !claims.IsForAudience("admin") {
			return fmt.Errorf("Invalid audience in JWT claims")
		}
	}

	if !claims.IsValidAt(time.Now()) {
		return fmt.Errorf("JWT token expired")
	}

	return nil
}

func (p *Local) Authorized(cookie, resource string) (string, error) {
	p.guard.Lock()
	secret := []byte(p.Key)
	p.guard.Unlock()

	verifier, err := jwt.NewVerifierHS(jwt.HS256, secret)
	if err != nil {
		return "", err
	}

	token, err := jwt.ParseAndVerifyString(cookie, verifier)
	if err != nil {
		return "", err
	}

	if err := verifier.Verify(token.Payload(), token.Signature()); err != nil {
		return "", err
	}

	var claims session
	if err := json.Unmarshal(token.RawClaims(), &claims); err != nil {
		return "", err
	}

	if !claims.IsForAudience("admin") {
		return "", fmt.Errorf("Invalid audience in JWT claims")
	}

	if !claims.IsValidAt(time.Now()) {
		return "", fmt.Errorf("JWT token expired")
	}

	if !p.authorised(claims.Role, resource) {
		return "", fmt.Errorf("%v not authorized for %s", claims.LoggedInAs, resource)
	}

	return claims.LoggedInAs, nil
}

func (p *Local) GetLoginId(cookie string) (*uuid.UUID, error) {
	p.guard.Lock()
	secret := []byte(p.Key)
	p.guard.Unlock()

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

	var claims login
	if err := json.Unmarshal(token.RawClaims(), &claims); err != nil {
		return nil, err
	}

	if !claims.IsForAudience("login") {
		return nil, fmt.Errorf("Invalid audience in JWT claims")
	}

	if !claims.IsValidAt(time.Now()) {
		return nil, fmt.Errorf("JWT token expired")
	}

	return &claims.LoginId, nil
}

func (p *Local) GetSessionId(cookie string) (*uuid.UUID, error) {
	p.guard.Lock()
	secret := []byte(p.Key)
	p.guard.Unlock()

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

	var claims session
	if err := json.Unmarshal(token.RawClaims(), &claims); err != nil {
		return nil, err
	}

	if !claims.IsForAudience("admin") {
		return nil, fmt.Errorf("Invalid audience in JWT claims")
	}

	if !claims.IsValidAt(time.Now()) {
		return nil, fmt.Errorf("JWT token expired")
	}

	return &claims.SessionId, nil
}

func (p *Local) authorised(role, resource string) bool {
	for _, r := range p.Resources {
		if r.Path.Match([]byte(resource)) && r.Authorised.Match([]byte(role)) {
			return true
		}
	}

	return false
}

func (p *Local) load(file string) error {
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	p.guard.Lock()
	defer p.guard.Unlock()
	if err := json.Unmarshal(bytes, p); err != nil {
		return err
	}

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
