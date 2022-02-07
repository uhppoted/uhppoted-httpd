package types

import (
	"time"

	core "github.com/uhppoted/uhppote-core/types"
)

type DateTime core.DateTime

func DateTimeNow() DateTime {
	return DateTime(core.DateTimeNow())
}

func (d DateTime) IsZero() bool {
	return core.DateTime(d).IsZero()
}

func (d DateTime) Before(t time.Time) bool {
	return core.DateTime(d).Before(t)
}

func (d DateTime) Add(dt time.Duration) DateTime {
	return DateTime(core.DateTime(d).Add(dt))
}

func (d DateTime) MarshalJSON() ([]byte, error) {
	return core.DateTime(d).MarshalJSON()
}

func (d *DateTime) UnmarshalJSON(bytes []byte) error {
	return ((*core.DateTime)(d)).UnmarshalJSON(bytes)
}

func (d DateTime) String() string {
	return core.DateTime(d).String()
}
