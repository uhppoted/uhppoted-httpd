package auth

import (
	"errors"
)

var ErrUnauthorised = errors.New("not authorised")
