package otp

import (
	"bytes"
	"image/png"

	"github.com/pquerna/otp/totp"

	"github.com/uhppoted/uhppoted-httpd/log"
)

func Get(uid, role string) ([]byte, error) {
	var b bytes.Buffer
	var options = totp.GenerateOpts{
		Issuer:      "uhppoted-httpd",
		AccountName: uid,
	}

	if key, err := totp.Generate(options); err != nil {
		return nil, err
	} else if img, err := key.Image(256, 256); err != nil {
		return nil, err
	} else {
		png.Encode(&b, img)

		return b.Bytes(), nil
	}
}

func warnf(format string, args ...any) {
	log.Warnf(format, args)
}
