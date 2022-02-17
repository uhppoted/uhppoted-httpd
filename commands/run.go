package commands

import (
	"flag"
	"fmt"
	"log"

	"github.com/uhppoted/uhppoted-httpd/audit"
	provider "github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/auth/impl"
	"github.com/uhppoted/uhppoted-httpd/httpd"
	"github.com/uhppoted/uhppoted-httpd/httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system"
	"github.com/uhppoted/uhppoted-lib/config"
)

type Run struct {
	console       bool
	configuration string
	debug         bool
}

func (r *Run) FlagSet() *flag.FlagSet {
	flagset := flag.NewFlagSet("", flag.ExitOnError)

	flagset.BoolVar(&r.console, "console", r.console, "Runs as a console application rather than a service")
	flagset.StringVar(&r.configuration, "config", r.configuration, "Sets the configuration file path")
	flagset.BoolVar(&r.debug, "debug", r.debug, "Enables detailed debugging logs")

	return flagset
}

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
		p, err := local.NewAuthProvider(conf.HTTPD.Security.AuthDB, conf.HTTPD.Security.LoginExpiry, conf.HTTPD.Security.SessionExpiry)
		if err != nil {
			return err
		}

		authentication, err = auth.NewBasic(p, conf.HTTPD.Security.AuthDB, conf.HTTPD.Security.CookieMaxAge)
		if err != nil {
			return err
		}
	}

	audit.SetAuditFile(conf.HTTPD.Audit.File)

	provider.Init(map[provider.RuleSet]string{
		provider.Interfaces:  conf.HTTPD.DB.Rules.System,
		provider.Controllers: conf.HTTPD.DB.Rules.System,
		provider.Doors:       conf.HTTPD.DB.Rules.Doors,
		provider.Cards:       conf.HTTPD.DB.Rules.Cards,
		provider.Groups:      conf.HTTPD.DB.Rules.Groups,
		provider.Events:      conf.HTTPD.DB.Rules.Events,
		provider.Logs:        conf.HTTPD.DB.Rules.Logs,
		provider.Users:       conf.HTTPD.DB.Rules.Users,
	})

	h := httpd.HTTPD{
		HTML:                     conf.HTTPD.HTML,
		AuthProvider:             authentication,
		HTTPEnabled:              conf.HTTPD.HttpEnabled,
		HTTPSEnabled:             conf.HTTPD.HttpsEnabled,
		CACertificate:            conf.HTTPD.CACertificate,
		TLSCertificate:           conf.HTTPD.TLSCertificate,
		TLSKey:                   conf.HTTPD.TLSKey,
		RequireClientCertificate: conf.HTTPD.RequireClientCertificate,
		RequestTimeout:           conf.HTTPD.RequestTimeout,
	}

	if err := system.Init(*conf, cmd.configuration, cmd.debug); err != nil {
		log.Fatalf("%5s Could not load system configuration (%v)", "FATAL", err)
	}

	h.Run()

	return nil
}
