package commands

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/uhppoted/uhppoted-api/config"
	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/httpd"
)

func (cmd *Run) Name() string {
	return "run"
}

func (cmd *Run) Description() string {
	return "Runs the uhppoted-httpd daemon/service until terminated by the system service manager"
}

func (cmd *Run) Usage() string {
	return "uhppoted-httpd [--debug] [--config <file>] [--logfile <file>] [--logfilesize <bytes>] [--pid <file>]"
}

func (cmd *Run) Help() {
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

func (cmd *Run) Execute(ctx context.Context) error {
	conf := config.NewConfig()
	if err := conf.Load(cmd.configuration); err != nil {
		log.Printf("%5s Could not load configuration (%v)", "WARN", err)
	}

	auth, err := auth.NewAuthProvider(conf.HTTPD.AuthDB, conf.HTTPD.SessionExpiry)
	if err != nil {
		return err
	}

	h := httpd.HTTPD{
		Dir:                      "html",
		AuthProvider:             auth,
		CookieMaxAge:             conf.HTTPD.CookieMaxAge,
		HTTPEnabled:              conf.HTTPD.HttpEnabled,
		HTTPSEnabled:             conf.HTTPD.HttpsEnabled,
		CACertificate:            conf.HTTPD.CACertificate,
		TLSCertificate:           conf.HTTPD.TLSCertificate,
		TLSKey:                   conf.HTTPD.TLSKey,
		RequireClientCertificate: conf.HTTPD.RequireClientCertificate,
	}

	h.Run()

	return nil
}
