package auth

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cristalhq/jwt/v3"
	"github.com/google/uuid"
)

var key = `secret`

type claims struct {
	jwt.StandardClaims
	LoggedInAs string `json:"uid"`
}

func Verify(cookie string) error {
	verifier, err := jwt.NewVerifierHS(jwt.HS256, []byte(key))
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

func Authorize(uid, pwd string) (string, error) {
	if uid != "admin" || pwd != "uhppoted" {
		return "", fmt.Errorf("Invalid login credentials")
	}

	signer, err := jwt.NewSignerHS(jwt.HS256, []byte(key))
	if err != nil {
		return "", err
	}

	uuid, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}

	expires := time.Now().Add(2 * time.Minute)

	claims := &claims{
		StandardClaims: jwt.StandardClaims{
			ID:        uuid.String(),
			Audience:  []string{"admin"},
			ExpiresAt: jwt.NewNumericDate(expires),
		},
		LoggedInAs: uid,
	}

	token, err := jwt.NewBuilder(signer).Build(claims)
	if err != nil {
		return "", err
	}

	return token.String(), nil
}
