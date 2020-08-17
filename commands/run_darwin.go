package commands

import (
	"flag"
)

type Run struct {
	console       bool
	configuration string
}

var RUN = Run{
	console:       false,
	configuration: "/usr/local/etc/com.github.uhppoted/uhppoted.conf",
}

func (r *Run) FlagSet() *flag.FlagSet {
	flagset := flag.NewFlagSet("", flag.ExitOnError)

	flagset.BoolVar(&r.console, "console", r.console, "Runs as a console application rather than a service")
	flagset.StringVar(&r.configuration, "config", r.configuration, "Sets the configuration file path")

	return flagset
}
