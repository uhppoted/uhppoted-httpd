package commands

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

var UNDAEMONIZE = Undaemonize{
	workdir: "/var/uhppoted",
	logdir:  "/var/log/uhppoted",
	config:  "/etc/uhppoted/uhppoted.conf",
}

type Undaemonize struct {
	workdir string
	logdir  string
	config  string
}

func (cmd *Undaemonize) Name() string {
	return "undaemonize"
}

func (cmd *Undaemonize) FlagSet() *flag.FlagSet {
	return flag.NewFlagSet("undaemonize", flag.ExitOnError)
}

func (cmd *Undaemonize) Description() string {
	return fmt.Sprintf("Deregisters %s as a service/daemon", SERVICE)
}

func (cmd *Undaemonize) Usage() string {
	return ""
}

func (cmd *Undaemonize) Help() {
	fmt.Println()
	fmt.Printf("  Usage: %s undaemonize\n", SERVICE)
	fmt.Println()
	fmt.Printf("    Deregisters %s from launchd as a service/daemon", SERVICE)
	fmt.Println()

	helpOptions(cmd.FlagSet())
}

func (cmd *Undaemonize) Execute(args ...interface{}) error {
	return fmt.Errorf("NOT IMPLEMENTED")
	//     fmt.Println("   ... undaemonizing")
	//
	//     if err := cmd.systemd(); err != nil {
	//         return err
	//     }
	//
	//     if err := cmd.logrotate(); err != nil {
	//         return err
	//     }
	//
	//     if err := cmd.clean(); err != nil {
	//         return err
	//     }
	//
	//     fmt.Printf("   ... %s unregistered as a systemd service\n", SERVICE)
	//     fmt.Printf(`
	//    NOTE: Configuration files in %s,
	//                working files in %s,
	//                and log files in %s
	//                were not removed and should be deleted manually
	// `, filepath.Dir(cmd.config), cmd.workdir, cmd.logdir)
	//     fmt.Println()
	//
	//     return nil
}

// func (cmd *Undaemonize) systemd() error {
//     path := filepath.Join("/etc/systemd/system", fmt.Sprintf("%s.service", SERVICE))
//     _, err := os.Stat(path)
//     if err != nil && !os.IsNotExist(err) {
//         return err
//     }
//
//     if os.IsNotExist(err) {
//         fmt.Printf("   ... nothing to do for 'systemd'   (%s does not exist)\n", path)
//         return nil
//     }
//
//     fmt.Printf("   ... stopping %s service\n", SERVICE)
//     command := exec.Command("systemctl", "stop", SERVICE)
//     out, err := command.CombinedOutput()
//     if strings.TrimSpace(string(out)) != "" {
//         fmt.Printf("   > %s\n", out)
//     }
//     if err != nil {
//         return fmt.Errorf("ERROR: Failed to stop '%s' (%v)\n", SERVICE, err)
//     }
//
//     fmt.Printf("   ... removing '%s'\n", path)
//     err = os.Remove(path)
//     if err != nil {
//         return err
//     }
//
//     return nil
// }
//
// func (cmd *Undaemonize) logrotate() error {
//     path := filepath.Join("/etc/logrotate.d", SERVICE)
//
//     fmt.Printf("   ... removing '%s'\n", path)
//
//     err := os.Remove(path)
//     if err != nil {
//         return err
//     }
//
//     return nil
// }
//
// func (cmd *Undaemonize) clean() error {
//     files := []string{
//         filepath.Join(cmd.workdir, fmt.Sprintf("%s.pid", SERVICE)),
//     }
//
//     directories := []string{
//         cmd.logdir,
//         cmd.workdir,
//     }
//
//     for _, f := range files {
//         fmt.Printf("   ... removing '%s'\n", f)
//         if err := os.Remove(f); err != nil && !os.IsNotExist(err) {
//             return err
//         }
//     }
//
//     for _, dir := range directories {
//         fmt.Printf("   ... removing '%s'\n", dir)
//         if err := os.Remove(dir); err != nil && !os.IsNotExist(err) {
//             patherr, ok := err.(*os.PathError)
//             if !ok {
//                 return err
//             }
//
//             syserr, ok := patherr.Err.(syscall.Errno)
//             if !ok {
//                 return err
//             }
//
//             if syserr != syscall.ENOTEMPTY {
//                 return err
//             }
//
//             fmt.Printf("   ... WARNING: could not remove directory '%s' (%v)\n", dir, syserr)
//         }
//     }
//
//     return nil
// }
