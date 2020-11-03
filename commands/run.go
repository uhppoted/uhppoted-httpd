package commands

import (
	"flag"
	"fmt"
	"log"

	"github.com/uhppoted/uhppoted-api/config"
	provider "github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/httpd"
	auth "github.com/uhppoted/uhppoted-httpd/httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/sys"
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

func (cmd *Run) Execute(args ...interface{}) error {
	conf := config.NewConfig()
	if err := conf.Load(cmd.configuration); err != nil {
		log.Printf("%5s Could not load configuration (%v)", "WARN", err)
	}

	var authentication auth.IAuth

	switch conf.HTTPD.Security.Auth {
	case "none":
		authentication = auth.NewNoneAuthenticator()

	default:
		p, err := provider.NewAuthProvider(conf.HTTPD.Security.AuthDB, conf.HTTPD.Security.LoginExpiry, conf.HTTPD.Security.SessionExpiry)
		if err != nil {
			return err
		}

		authentication = auth.NewBasicAuthenticator(p, conf.HTTPD.Security.CookieMaxAge, conf.HTTPD.Security.StaleTime)
	}

	h := httpd.HTTPD{
		Dir:                      "html",
		AuthProvider:             authentication,
		HTTPEnabled:              conf.HTTPD.HttpEnabled,
		HTTPSEnabled:             conf.HTTPD.HttpsEnabled,
		CACertificate:            conf.HTTPD.CACertificate,
		TLSCertificate:           conf.HTTPD.TLSCertificate,
		TLSKey:                   conf.HTTPD.TLSKey,
		RequireClientCertificate: conf.HTTPD.RequireClientCertificate,
		RequestTimeout:           conf.HTTPD.RequestTimeout,
		DB: struct {
			File  string
			Rules string
		}{
			File:  conf.HTTPD.DB.File,
			Rules: conf.HTTPD.DB.Rules,
		},
		Audit: struct {
			File string
		}{
			File: conf.HTTPD.Audit.File,
		},
	}

	if err := system.Init(conf.HTTPD.System.File); err != nil {
		log.Fatalf("%5s Could not load system configuration (%v)", "FATAL", err)
	}

	h.Run()

	return nil
}
