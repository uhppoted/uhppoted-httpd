package types

type Status int

const (
	StatusUnknown Status = iota
	StatusOk
	StatusUncertain
	StatusError
	StatusUnconfigured
	StatusNew
	StatusDeleted
	StatusIncomplete
)

func (s Status) String() string {
	return [...]string{"unknown", "ok", "uncertain", "error", "unconfigured", "new", "deleted", "incomplete"}[s]
}
