package types

type HttpdError struct {
	Status int
	Err    error
	Detail error
}

func (e *HttpdError) Error() string {
	return e.Err.Error()
}
