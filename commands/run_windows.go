package commands

import (
	"path/filepath"
)

var RUN = Run{
	console:       false,
	configuration: filepath.Join(workdir(), "uhppoted.conf"),
}
