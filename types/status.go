package types

type Status int

const (
	StatusUnknown Status = iota
	StatusOk
	StatusUncertain
	StatusError
	StatusUnconfigured
)

func (s Status) String() string {
	return [...]string{"unknown", "ok", "uncertain", "error", "unconfigured"}[s]
}
