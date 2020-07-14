package commands

import (
	"flag"
)

type Run struct {
	configuration string
}

var RUN = Run{
	configuration: "/usr/local/etc/com.github.twystd.uhppoted/uhppoted.conf",
}

func (r *Run) FlagSet() *flag.FlagSet {
	flagset := flag.NewFlagSet("", flag.ExitOnError)

	flagset.StringVar(&r.configuration, "config", r.configuration, "Sets the configuration file path")

	return flagset
}
