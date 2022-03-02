package commands

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
)

var UNDAEMONIZE = Undaemonize{
	name:    SERVICE,
	workdir: workdir(),
	logdir:  filepath.Join(workdir(), "logs"),
	config:  workdir(),
	etc:     filepath.Join(workdir(), "httpd"),
}

type Undaemonize struct {
	name    string
	workdir string
	logdir  string
	config  string
	etc     string
}

// Ref. https://docs.microsoft.com/en-us/windows/win32/debug/system-error-codes--1000-1299-
const ERROR_SERVICE_NOT_ACTIVE = 0x426

func (cmd *Undaemonize) Name() string {
	return "undaemonize"
}

func (cmd *Undaemonize) FlagSet() *flag.FlagSet {
	return flag.NewFlagSet("undaemonize", flag.ExitOnError)
}

func (cmd *Undaemonize) Description() string {
	return fmt.Sprintf("Deregisters %s from the list of Windows services", SERVICE)
}

func (cmd *Undaemonize) Usage() string {
	return ""
}

func (cmd *Undaemonize) Help() {
	fmt.Println()
	fmt.Printf("  Usage: %s undaemonize\n", SERVICE)
	fmt.Println()
	fmt.Printf("    Deregisters %s from the list of Windows services", SERVICE)
	fmt.Println()

	helpOptions(cmd.FlagSet())
}

func (cmd *Undaemonize) Execute(args ...interface{}) error {
	fmt.Println("   ... undaemonizing")

	if err := cmd.unregister(); err != nil {
		return err
	}

	if err := cmd.clean(); err != nil {
		return err
	}

	fmt.Printf("   ... %s deregistered as a Windows service\n", SERVICE)
	fmt.Printf(`
       NOTE: Configuration files in %s,
             working files in %s,
             log files in %s,
             HTML files in %s,
             and GRULES files in %s
             were not removed and should be deleted manually
`, filepath.Dir(cmd.config), cmd.workdir, cmd.logdir, filepath.Join(cmd.etc, "html"), filepath.Join(cmd.etc, "grules"))
	fmt.Println()

	return nil
}

func (cmd *Undaemonize) unregister() error {
	fmt.Printf("   ... unregistering %s as a Windows service\n", cmd.name)
	m, err := mgr.Connect()
	if err != nil {
		return err
	}

	defer m.Disconnect()

	s, err := m.OpenService(cmd.name)
	if err != nil {
		return fmt.Errorf("service %s is not installed", cmd.name)
	}

	defer s.Close()

	fmt.Printf("   ... stopping %s service\n", cmd.name)
	status, err := s.Control(svc.Stop)
	if err != nil {
		// Ref. https://stackoverflow.com/questions/63470776/how-to-get-windows-system-error-code-when-calling-windows-api-in-go
		if syserr, ok := err.(syscall.Errno); ok {
			if syserr != ERROR_SERVICE_NOT_ACTIVE {
				return err
			}
		}
	} else {
		fmt.Printf("   ... %s stopped: %v\n", cmd.name, status)
	}

	fmt.Printf("   ... deleting %s service\n", cmd.name)
	err = s.Delete()
	if err != nil {
		return err
	}

	err = eventlog.Remove(cmd.name)
	if err != nil {
		return fmt.Errorf("RemoveEventLogSource() failed: %s", err)
	}

	fmt.Printf("   ... %s unregistered from the list of Windows services\n", cmd.name)
	return nil
}

func (cmd *Undaemonize) clean() error {
	files := []string{
		filepath.Join(cmd.workdir, fmt.Sprintf("%s.pid", SERVICE)),
	}

	directories := []string{
		cmd.logdir,
		cmd.workdir,
	}

	for _, f := range files {
		fmt.Printf("   ... removing '%s'\n", f)
		if err := os.Remove(f); err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	warnings := []string{}
	for _, dir := range directories {
		fmt.Printf("   ... removing '%s'\n", dir)
		if err := os.Remove(dir); err != nil && !os.IsNotExist(err) {
			patherr, ok := err.(*os.PathError)
			if !ok {
				return err
			}

			syserr, ok := patherr.Err.(syscall.Errno)
			if !ok {
				return err
			}

			// Windows error is: ERROR_DIR_NOT_EMPTY (0x91). May be fixed in 1.14.
			// Ref. https://github.com/golang/go/issues/32309
			// Ref. https://docs.microsoft.com/en-us/windows/win32/debug/system-error-codes--0-499-
			if syserr != syscall.ENOTEMPTY && syserr != 0x91 {
				return err
			}

			warnings = append(warnings, fmt.Sprintf("could not remove directory '%s' (%v)", dir, syserr))
		}
	}

	if len(warnings) > 0 {
		fmt.Println()
		for _, w := range warnings {
			fmt.Printf("   ... WARNING: %v\n", w)
		}
	}

	return nil
}
