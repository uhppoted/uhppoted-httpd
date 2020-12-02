package system

type status int

const (
	StatusUnknown status = iota
	StatusOk
	StatusUncertain
	StatusError
)

func (s status) String() string {
	return [...]string{"unknown", "ok", "uncertain", "error"}[s]
}
