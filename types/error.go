package types

import (
	"net/http"
)

type HttpdError struct {
	Status int
	Err    error
	Detail error
}

func (e *HttpdError) Error() string {
	return e.Err.Error()
}

func Unauthorised(msg, err error) *HttpdError {
	return &HttpdError{
		Status: http.StatusUnauthorized,
		Err:    msg,
		Detail: err,
	}
}

func BadRequest(msg, err error) *HttpdError {
	return &HttpdError{
		Status: http.StatusBadRequest,
		Err:    msg,
		Detail: err,
	}
}
