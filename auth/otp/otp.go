package otp

import (
	"bytes"
	"fmt"
	"image/png"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	lib "github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"github.com/uhppoted/uhppoted-httpd/system"
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

func Get(uid, role, keyid string) (string, time.Duration, []byte, error) {
	var secret *otpkey
	var now = time.Now()
	var options = totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: uid,
		Period:      30,
		Algorithm:   lib.AlgorithmSHA1,
		Digits:      lib.DigitsSix,
	}

	if k, ok := secrets[keyid]; ok && k != nil && k.expires.After(now) {
		secret = k
	} else if u, ok := system.GetUser(uid); !ok || u == nil {
		return "", 0, nil, fmt.Errorf("invalid login credentials")
	} else if key := u.OTPKey(); key != "" {
		v := url.Values{}

		v.Set("issuer", options.Issuer)
		v.Set("period", strconv.FormatUint(uint64(options.Period), 10))
		v.Set("algorithm", options.Algorithm.String())
		v.Set("digits", options.Digits.String())
		v.Set("secret", key)

		u := url.URL{
			Scheme:   "otpauth",
			Host:     "totp",
			Path:     "/" + options.Issuer + ":" + options.AccountName,
			RawQuery: v.Encode(),
		}

		if k, err := lib.NewKeyFromURL(u.String()); err != nil {
			return "", 0, nil, err
		} else if uuid, err := uuid.NewUUID(); err != nil {
			return "", 0, nil, err
		} else {
			keyid = fmt.Sprintf("%v", uuid)
			secret = &otpkey{
				key:     k,
				expires: time.Now().Add(5 * time.Minute),
			}

			secrets[keyid] = secret
		}
	} else {
		if key, err := totp.Generate(options); err != nil {
			return "", 0, nil, err
		} else if key == nil {
			return "", 0, nil, fmt.Errorf("invalid OTP key")
		} else if uuid, err := uuid.NewUUID(); err != nil {
			return "", 0, nil, err
		} else {
			keyid = fmt.Sprintf("%v", uuid)
			secret = &otpkey{
				key:     key,
				expires: time.Now().Add(1 * time.Minute),
			}

			secrets[keyid] = secret
		}
	}

	var b bytes.Buffer

	if img, err := secret.key.Image(256, 256); err != nil {
		return "", 0, nil, err
	} else {
		secrets[keyid].expires = time.Now().Add(5 * time.Minute)

		png.Encode(&b, img)

		return keyid, 2 * time.Minute, b.Bytes(), nil
	}
}

func Validate(uid string, role string, keyid string, otp string) error {
	now := time.Now()

	if secret, ok := secrets[keyid]; !ok || secret == nil || !secret.expires.After(now) {
		return fmt.Errorf("invalid OTP secret")
	} else if !totp.Validate(otp, secret.key.Secret()) {
		return fmt.Errorf("invalid OTP")
	} else if err := system.SetOTP(uid, role, secret.key.Secret()); err != nil {
		return err
	}

	return nil
}

func Revoke(uid, role string) error {
	return system.RevokeOTP(uid, role)
}

func Verify(uid string, role string, otp string) bool {
	if u, ok := system.GetUser(uid); !ok || u == nil {
		return false
	} else if secret := u.OTPKey(); secret == "" {
		return false
	} else {
		return totp.Validate(otp, secret)
	}
}

func Enabled(uid, role string) bool {
	if u, ok := system.GetUser(uid); !ok || u == nil {
		return false
	} else if secret := u.OTPKey(); secret == "" {
		return false
	} else {
		return regexp.MustCompile("[ABCDEFGHIJKLMNOPQRSTUVWXYZ234567]{16,}").MatchString(secret)
	}
}
