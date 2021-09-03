package commands

import (
	"flag"
	"fmt"
	"log"

	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/hyperjumptech/grule-rule-engine/builder"
	"github.com/hyperjumptech/grule-rule-engine/pkg"

	"github.com/uhppoted/uhppoted-httpd/audit"
	provider "github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/httpd"
	auth "github.com/uhppoted/uhppoted-httpd/httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system"
	"github.com/uhppoted/uhppoted-httpd/system/cards/grule"
	"github.com/uhppoted/uhppoted-lib/config"
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

		authentication = auth.NewBasicAuthenticator(p, conf.HTTPD.Security.CookieMaxAge, conf.HTTPD.Security.StaleTime, []string{"/authenticate"})
	}

	trail := audit.NewAuditTrail(conf.HTTPD.Audit.File)

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
			GRules struct {
				System string
				Cards  string
				Doors  string
				Groups string
			}
		}{
			GRules: struct {
				System string
				Cards  string
				Doors  string
				Groups string
			}{
				System: conf.HTTPD.DB.Rules.System,
				Cards:  conf.HTTPD.DB.Rules.Cards,
				Doors:  conf.HTTPD.DB.Rules.Doors,
				Groups: conf.HTTPD.DB.Rules.Groups,
			},
		},

		Audit: trail,
	}

	permissions := ast.NewKnowledgeLibrary()
	if err := builder.NewRuleBuilder(permissions).BuildRuleFromResource("acl", "0.0.0", pkg.NewFileResource(conf.HTTPD.DB.Rules.ACL)); err != nil {
		log.Fatal(fmt.Errorf("Error loading ACL ruleset (%v)", err))
	}

	ruleset, err := grule.NewGrule(permissions)
	if err != nil {
		log.Fatal(fmt.Errorf("Error initialising ACL ruleset (%v)", err))
	}

	if err := system.Init(*conf, cmd.configuration, ruleset, trail); err != nil {
		log.Fatalf("%5s Could not load system configuration (%v)", "FATAL", err)
	}

	h.Run()

	return nil
}
