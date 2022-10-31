package otp

import (
	"bytes"
	"fmt"
	"image/png"
	"strings"
	"time"

	"github.com/google/uuid"
	lib "github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"github.com/uhppoted/uhppoted-httpd/log"
)

type otpkey struct {
	key     *lib.Key
	expires time.Time
}

var issuer = "uhppoted-httpd"
var secrets = map[string]*otpkey{}

func SetIssuer(u string) {
	if v := strings.TrimSpace(u); v != "" {
		issuer = v
	}
}

func Get(uid, role, keyid string) (string, []byte, error) {
	var b bytes.Buffer
	var options = totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: uid,
	}

	var secret *otpkey
	var now = time.Now()

	if k, ok := secrets[keyid]; ok && k != nil && k.expires.After(now) {
		secret = k
	} else {
		if key, err := totp.Generate(options); err != nil {
			return "", nil, err
		} else if key == nil {
			return "", nil, fmt.Errorf("invalid OTP key")
		} else if uuid, err := uuid.NewUUID(); err != nil {
			return "", nil, err
		} else {
			keyid = fmt.Sprintf("%v", uuid)
			secret = &otpkey{
				key:     key,
				expires: time.Now().Add(1 * time.Minute),
			}

			secrets[keyid] = secret
		}
	}

	if img, err := secret.key.Image(256, 256); err != nil {
		return "", nil, err
	} else {
		png.Encode(&b, img)

		return keyid, b.Bytes(), nil
	}
}

func Validate(keyid string, otps ...string) error {
	var now = time.Now()

	if len(otps) == 0 {
	} else if secret, ok := secrets[keyid]; !ok || secret == nil || !secret.expires.After(now) {
		return fmt.Errorf("Invalid OTP secret")
	} else {
		for _, otp := range otps {
			if !totp.Validate(otp, secret.key.Secret()) {
				return fmt.Errorf("Invalid OTP")
			}
		}
	}

	return nil
}

func warnf(format string, args ...any) {
	log.Warnf(format, args...)
}
