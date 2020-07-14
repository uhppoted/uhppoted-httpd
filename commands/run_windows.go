package commands

import (
	"flag"
	"path/filepath"
)

type Run struct {
	configuration string
}

var RUN = Run{
	configuration: filepath.Join(workdir(), "uhppoted.conf"),
}

func (r *Run) FlagSet() *flag.FlagSet {
	flagset := flag.NewFlagSet("", flag.ExitOnError)

	flagset.StringVar(&r.configuration, "config", r.configuration, "Sets the configuration file path")

	return flagset
}
