package commands

import (
	"context"
	"flag"
	"fmt"
)

type RunCmd struct {
}

var RUN = RunCmd{}

func (cmd *RunCmd) Name() string {
	return "run"
}

func (cmd *RunCmd) Description() string {
	return "Runs the uhppoted-httpd daemon/service until terminated by the system service manager"
}

func (cmd *RunCmd) Usage() string {
	return "uhppoted-httpd [--debug] [--config <file>] [--logfile <file>] [--logfilesize <bytes>] [--pid <file>]"
}

func (cmd *RunCmd) Help() {
	fmt.Println()
	fmt.Println("  Usage: uhppoted-httpd <options>")
	fmt.Println()
	fmt.Println("  Options:")
	fmt.Println()
	cmd.FlagSet().VisitAll(func(f *flag.Flag) {
		fmt.Printf("    --%-12s %s\n", f.Name, f.Usage)
	})
	fmt.Println()
}

func (cmd *RunCmd) FlagSet() *flag.FlagSet {
	return flag.NewFlagSet("run", flag.ExitOnError)
}

func (cmd *RunCmd) Execute(ctx context.Context) error {
	fmt.Println("   ... running (sort of)")

	return fmt.Errorf("NOT IMPLEMENTED")
}
