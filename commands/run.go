package commands

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/uhppoted/uhppoted-lib/config"
	"github.com/uhppoted/uhppoted-lib/lockfile"

	"github.com/uhppoted/uhppoted-httpd/audit"
	provider "github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/auth/impl"
	"github.com/uhppoted/uhppoted-httpd/auth/otp"
	"github.com/uhppoted/uhppoted-httpd/httpd"
	"github.com/uhppoted/uhppoted-httpd/httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/log"
	"github.com/uhppoted/uhppoted-httpd/system"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type Run struct {
	console       bool
	mode          string
	configuration string
	debug         bool
	workdir       string
	lockfile      string
	logFile       string
	logFileSize   int
}

func (r *Run) FlagSet() *flag.FlagSet {
	flagset := flag.NewFlagSet("", flag.ExitOnError)
	lockfile := filepath.Join(os.TempDir(), fmt.Sprintf("%s.pid", SERVICE))

	flagset.BoolVar(&r.console, "console", false, "Runs as a console application rather than a service")
	flagset.StringVar(&r.mode, "mode", "update", "Sets the run mode (normal/monitor/synchronize). Defaults to 'normal'")
	flagset.StringVar(&r.configuration, "config", r.configuration, "Sets the configuration file path")
	flagset.StringVar(&r.lockfile, "lockfile", r.lockfile, fmt.Sprintf("(optional) lockfile used to prevent running multiple copies of the service. Defaults to %v", lockfile))
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
		log.Warnf("Could not load configuration (%v)", err)
	}

	// ... initialise timezones

	if conf.HTTPD.Timezones != "" {
		types.LoadTimezones(conf.HTTPD.Timezones)
	}

	// ... panic handler
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("%v\n", err)
			os.Exit(-1)
		}
	}()

	// ... create lockfile
	lockFile := config.Lockfile{
		File:   cmd.lockfile,
		Remove: conf.LockfileRemove,
	}

	if lockFile.File == "" {
		lockFile.File = filepath.Join(os.TempDir(), fmt.Sprintf("%s.pid", SERVICE))
	}

	if lock, err := lockfile.MakeLockFile(lockFile); err != nil {
		return err
	} else {
		defer func() {
			lock.Release()
		}()

		log.SetFatalHook(func() {
			lock.Release()
		})
	}

	// ... cleanup dangling temporary files

	cleanup(*conf)

	// 'k, good to go
	f(*conf)

	return nil
}

func (cmd *Run) run(conf config.Config, interrupt chan os.Signal) {
	// ... initialise auth providers
	var authentication auth.IAuth

	switch conf.HTTPD.Security.Auth {
	case "none":
		authentication = auth.NewNoneAuthenticator()

	default:
		p, err := local.NewAuthProvider(
			conf.HTTPD.Security.AuthDB,
			conf.HTTPD.Security.LoginExpiry,
			conf.HTTPD.Security.SessionExpiry,
			conf.HTTPD.Security.OTP.Login == "allow")
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

	// ... initialise OTP

	otp.SetIssuer(conf.Security.OTP.Issuer)

	// ... run
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
		panic(fmt.Errorf("could not load system configuration (%v)", err))
	}

	h.Run(runMode, conf.HTTPD.PIN.Enabled, interrupt)
}

func cleanup(cfg config.Config) {
	folders := map[string]struct{}{}
	files := []string{
		cfg.HTTPD.System.Interfaces,
		cfg.HTTPD.System.Controllers,
		cfg.HTTPD.System.Doors,
		cfg.HTTPD.System.Cards,
		cfg.HTTPD.System.Groups,
		cfg.HTTPD.System.Events,
		cfg.HTTPD.System.Logs,
		cfg.HTTPD.System.Users,
		cfg.HTTPD.System.History,
	}

	for _, file := range files {
		dir := filepath.Dir(file)
		folders[dir] = struct{}{}
	}

	for dir := range folders {
		glob := filepath.Join(dir, "*.tmp")

		if tempfiles, err := filepath.Glob(glob); err != nil {
			log.Warnf("%v", err)
		} else {
			for _, file := range tempfiles {
				if err := os.Remove(file); err != nil {
					log.Warnf("Error deleting leftover temporary file  %v (%v)", file, err)
				} else {
					log.Infof("Deleting leftover temporary file  %v", file)
				}
			}
		}
	}

}
