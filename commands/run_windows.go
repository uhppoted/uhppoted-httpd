package commands

import (
	"flag"
	"path/filepath"
)

type Run struct {
	console       bool
	configuration string
}

var RUN = Run{
	console:       false,
	configuration: filepath.Join(workdir(), "uhppoted.conf"),
}

func (r *Run) FlagSet() *flag.FlagSet {
	flagset := flag.NewFlagSet("", flag.ExitOnError)

	flagset.BoolVar(&r.console, "console", r.console, "Runs as a console application rather than a service")
	flagset.StringVar(&r.configuration, "config", r.configuration, "Sets the configuration file path")

	return flagset
}
