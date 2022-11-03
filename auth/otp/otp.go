package otp

import (
	"bytes"
	"fmt"
	"image/png"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	lib "github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"github.com/uhppoted/uhppoted-httpd/log"
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

func Get(uid, keyid string) (string, []byte, error) {
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
	} else if key, err := system.GetOTP(uid); err != nil {
		return "", nil, err
	} else if key != "" {
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
			return "", nil, err
		} else if uuid, err := uuid.NewUUID(); err != nil {
			return "", nil, err
		} else {
			keyid = fmt.Sprintf("%v", uuid)
			secret = &otpkey{
				key:     k,
				expires: time.Now().Add(1 * time.Minute),
			}

			secrets[keyid] = secret
		}
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

	var b bytes.Buffer

	if img, err := secret.key.Image(256, 256); err != nil {
		return "", nil, err
	} else {
		png.Encode(&b, img)

		return keyid, b.Bytes(), nil
	}
}

func Validate(uid string, keyid string, otp string) error {
	now := time.Now()

	if secret, ok := secrets[keyid]; !ok || secret == nil || !secret.expires.After(now) {
		return fmt.Errorf("Invalid OTP secret")
	} else if !totp.Validate(otp, secret.key.Secret()) {
		return fmt.Errorf("Invalid OTP")
	} else if err := system.SetOTP(uid, secret.key.Secret()); err != nil {
		return err
	}

	return nil
}

func Verify(uid string, otp string) bool {
	if secret, err := system.GetOTP(uid); err == nil {
		return totp.Validate(otp, secret)
	}

	return false
}

func warnf(format string, args ...any) {
	log.Warnf(format, args...)
}
