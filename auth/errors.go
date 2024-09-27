package auth

import (
	"errors"
)

var ErrUnauthorised = errors.New("not authorised")
var ErrDoNotCache = errors.New("not cacheable")
