package commands

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/uhppoted/uhppoted-lib/config"

	"github.com/uhppoted/uhppoted-httpd/audit"
	provider "github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/auth/impl"
	"github.com/uhppoted/uhppoted-httpd/httpd"
	"github.com/uhppoted/uhppoted-httpd/httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/system"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type Run struct {
	console       bool
	mode          string
	configuration string
	debug         bool
	workdir       string
	logFile       string
	logFileSize   int
}

func (r *Run) FlagSet() *flag.FlagSet {
	flagset := flag.NewFlagSet("", flag.ExitOnError)

	flagset.BoolVar(&r.console, "console", false, "Runs as a console application rather than a service")
	flagset.StringVar(&r.mode, "mode", "update", "Sets the run mode (normal/monitor/synchronize). Defaults to 'normal'")
	flagset.StringVar(&r.configuration, "config", r.configuration, "Sets the configuration file path")
	flagset.BoolVar(&r.debug, "debug", false, "Enables detailed debugging logs")

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

func (cmd *Run) execute(f func(c config.Config)) error {
	// ... load configuration
	conf := config.NewConfig()
	if err := conf.Load(cmd.configuration); err != nil {
		log.Printf("%5s Could not load configuration (%v)", "WARN", err)
	}

	// ... initialise timezones

	if conf.HTTPD.Timezones != "" {
		types.LoadTimezones(conf.HTTPD.Timezones)
	}

	// ... create lockfile
	if err := os.MkdirAll(cmd.workdir, os.ModeDir|os.ModePerm); err != nil {
		return fmt.Errorf("Unable to create working directory '%v': %v", cmd.workdir, err)
	}

	pid := fmt.Sprintf("%d\n", os.Getpid())
	lockfile := filepath.Join(cmd.workdir, fmt.Sprintf("%s.pid", SERVICE))

	if _, err := os.Stat(lockfile); err == nil {
		return fmt.Errorf("PID lockfile '%v' already in use", lockfile)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("Error checking PID lockfile '%v' (%v)", lockfile, err)
	}

	if err := os.WriteFile(lockfile, []byte(pid), 0644); err != nil {
		return fmt.Errorf("Unable to create PID lockfile: %v", err)
	}

	defer func() {
		if err := recover(); err != nil {
			log.Printf("%-5s %v\n", "FATAL", err)
			os.Exit(-1)
		}
	}()

	defer os.Remove(lockfile)

	f(*conf)

	return nil
}

func (cmd *Run) run(conf config.Config, interrupt chan os.Signal) {
	var authentication auth.IAuth

	switch conf.HTTPD.Security.Auth {
	case "none":
		authentication = auth.NewNoneAuthenticator()

	default:
		p, err := local.NewAuthProvider(conf.HTTPD.Security.AuthDB, conf.HTTPD.Security.LoginExpiry, conf.HTTPD.Security.SessionExpiry)
		if err != nil {
			panic(fmt.Sprintf("Error instantiating auth provider (%v)", err))
		}

		authentication, err = auth.NewBasic(p, conf.HTTPD.Security.AuthDB, conf.HTTPD.Security.CookieMaxAge)
		if err != nil {
			panic(fmt.Sprintf("Error instantiating 'basic' auth provider (%v)", err))
		}
	}

	if err := audit.SetAuditFile(conf.HTTPD.Audit.File); err != nil {
		panic(err)
	}

	provider.Init(map[provider.RuleSet]string{
		provider.Interfaces:  conf.HTTPD.DB.Rules.Interfaces,
		provider.Controllers: conf.HTTPD.DB.Rules.Controllers,
		provider.Doors:       conf.HTTPD.DB.Rules.Doors,
		provider.Cards:       conf.HTTPD.DB.Rules.Cards,
		provider.Groups:      conf.HTTPD.DB.Rules.Groups,
		provider.Events:      conf.HTTPD.DB.Rules.Events,
		provider.Logs:        conf.HTTPD.DB.Rules.Logs,
		provider.Users:       conf.HTTPD.DB.Rules.Users,
	})

	h := httpd.HTTPD{
		HTML:                     conf.HTTPD.HTML,
		HttpEnabled:              conf.HTTPD.HttpEnabled,
		HttpsEnabled:             conf.HTTPD.HttpsEnabled,
		HttpPort:                 conf.HTTPD.HttpPort,
		HttpsPort:                conf.HTTPD.HttpsPort,
		AuthProvider:             authentication,
		CACertificate:            conf.HTTPD.CACertificate,
		TLSCertificate:           conf.HTTPD.TLSCertificate,
		TLSKey:                   conf.HTTPD.TLSKey,
		RequireClientCertificate: conf.HTTPD.RequireClientCertificate,
		RequestTimeout:           conf.HTTPD.RequestTimeout,
	}

	runMode := types.ParseRunMode(cmd.mode)

	if err := system.Init(conf, cmd.configuration, runMode, cmd.debug); err != nil {
		panic(fmt.Errorf("Could not load system configuration (%v)", err))
	}

	h.Run(runMode, interrupt)
}
