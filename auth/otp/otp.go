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
var keys = map[string]*otpkey{}

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

	if k, ok := keys[keyid]; ok && k != nil && k.expires.After(now) {
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

			keys[keyid] = secret
		}
	}

	if img, err := secret.key.Image(256, 256); err != nil {
		return "", nil, err
	} else {
		png.Encode(&b, img)

		return keyid, b.Bytes(), nil
	}
}

func warnf(format string, args ...any) {
	log.Warnf(format, args)
}
