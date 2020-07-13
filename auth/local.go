package auth

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"

	"github.com/cristalhq/jwt/v3"
	"github.com/google/uuid"
)

type Local struct {
	Key         string           `json:"key"`
	Users       map[string]*user `json:"users"`
	TokenExpiry expiry           `json:"token-expiry"`

	guard sync.Mutex
}

type claims struct {
	jwt.StandardClaims
	LoggedInAs string `json:"uid"`
}

type user struct {
	Password string `json:"password"`
}

type expiry time.Duration

func (x *expiry) UnmarshalJSON(bytes []byte) error {
	var s string

	err := json.Unmarshal(bytes, &s)
	if err != nil {
		return err
	}

	dt, err := time.ParseDuration(s)
	if err != nil {
		return err
	}

	*x = expiry(dt)

	return nil
}

func NewLocalAuthProvider(file string) (*Local, error) {
	provider := Local{}
	if err := provider.load(file); err != nil {
		return nil, err
	}

	provider.watch(file)

	return &provider, nil
}

func (p *Local) Authorize(uid, pwd string) (string, error) {
	p.guard.Lock()
	secret := []byte(p.Key)
	users := p.Users
	expiry := p.TokenExpiry
	p.guard.Unlock()

	hash := fmt.Sprintf("%0x", sha256.Sum256([]byte(pwd)))

	if u, ok := users[uid]; !ok {
		return "", fmt.Errorf("Invalid login credentials")
	} else if hash != u.Password {
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

	claims := &claims{
		StandardClaims: jwt.StandardClaims{
			ID:        uuid.String(),
			Audience:  []string{"admin"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expiry))),
		},
		LoggedInAs: uid,
	}

	token, err := jwt.NewBuilder(signer).Build(claims)
	if err != nil {
		return "", err
	}

	return token.String(), nil
}

func (p *Local) Verify(cookie string) error {
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

	var claims claims
	if err := json.Unmarshal(token.RawClaims(), &claims); err != nil {
		return err
	}

	if !claims.IsForAudience("admin") {
		return fmt.Errorf("Invalid audience in JWT claims")
	}

	if !claims.IsValidAt(time.Now()) {
		return fmt.Errorf("JWT token expired")
	}

	return nil
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

// NOTE: interim file watcher implementation pending fsnotify in Go 1.4
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
