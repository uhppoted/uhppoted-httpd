package doors

type status int

const (
	StatusUnknown status = iota
	StatusOk
	StatusUncertain
	StatusError
	StatusUnconfigured
	StatusNew
	StatusDeleted
)

func (s status) String() string {
	return [...]string{"unknown", "ok", "uncertain", "error", "unconfigured", "new", "deleted"}[s]
}
