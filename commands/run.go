package commands

import (
	"context"
	"flag"
	"fmt"

	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/httpd"
)

type RunCmd struct {
	AuthDB string
}

var RUN = RunCmd{
	AuthDB: "/usr/local/etc/com.github.twystd.uhppoted/httpd/users.json",
}

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
	auth, err := auth.NewAuthProvider(cmd.AuthDB)
	if err != nil {
		return err
	}

	h := httpd.HTTPD{
		Dir:          "html",
		AuthProvider: auth,
	}

	h.Run()

	return nil
}
