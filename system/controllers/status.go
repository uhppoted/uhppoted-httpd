package controllers

type status int

const (
	StatusUnknown status = iota
	StatusOk
	StatusUncertain
	StatusError
	StatusUnconfigured
	StatusNew
)

func (s status) String() string {
	return [...]string{"unknown", "ok", "uncertain", "error", "unconfigured", "new"}[s]
}
