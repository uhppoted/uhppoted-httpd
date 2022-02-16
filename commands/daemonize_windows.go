package commands

import (
	//	"bufio"
	"flag"
	"fmt"
	//	"os"
	"path/filepath"
	//	"strings"

	//	"golang.org/x/sys/windows/svc/eventlog"
	//	"golang.org/x/sys/windows/svc/mgr"

	"github.com/uhppoted/uhppote-core/types"
	//	"github.com/uhppoted/uhppoted-lib/config"
)

var DAEMONIZE = Daemonize{
	name:        SERVICE,
	description: "UHPPOTE UTO311-L0x access card controllers service",
	workdir:     workdir(),
	logdir:      filepath.Join(workdir(), "logs"),
	config:      filepath.Join(workdir(), "uhppoted.conf"),
	hotp:        filepath.Join(workdir(), "mqtt.hotp.secrets"),
}

type info struct {
	Executable       string
	WorkDir          string
	LogDir           string
	BindAddress      *types.BindAddr
	BroadcastAddress *types.BroadcastAddr
}

type Daemonize struct {
	name        string
	description string
	workdir     string
	logdir      string
	config      string
	hotp        string
}

func (cmd *Daemonize) Name() string {
	return "daemonize"
}

func (cmd *Daemonize) FlagSet() *flag.FlagSet {
	return flag.NewFlagSet("daemonize", flag.ExitOnError)
}

func (cmd *Daemonize) Description() string {
	return fmt.Sprintf("Registers %s as a Windows service", SERVICE)
}

func (cmd *Daemonize) Usage() string {
	return ""
}

func (cmd *Daemonize) Help() {
	fmt.Println()
	fmt.Printf("  Usage: %s daemonize\n", SERVICE)
	fmt.Println()
	fmt.Printf("    Registers %s as a Windows service\n", SERVICE)
	fmt.Println()

	helpOptions(cmd.FlagSet())
}

func (cmd *Daemonize) Execute(args ...interface{}) error {
	return fmt.Errorf("NOT IMPLEMENTED")
	//    dir := filepath.Dir(cmd.config)
	//    r := bufio.NewReader(os.Stdin)
	//
	//    fmt.Println()
	//    fmt.Printf("     **** PLEASE MAKE SURE YOU HAVE A BACKUP COPY OF THE CONFIGURATION INFORMATION AND KEYS IN %s ***\n", dir)
	//    fmt.Println()
	//    fmt.Printf("     Enter 'yes' to continue with the installation: ")
	//
	//    text, err := r.ReadString('\n')
	//    if err != nil || strings.TrimSpace(text) != "yes" {
	//        fmt.Println()
	//        fmt.Printf("     -- installation cancelled --")
	//        fmt.Println()
	//        return nil
	//    }
	//
	//    return cmd.execute()
}

// func (cmd *Daemonize) execute() error {
//     fmt.Println()
//     fmt.Println("   ... daemonizing")
//
//     executable, err := os.Executable()
//     if err != nil {
//         return err
//     }
//
//     bind, broadcast, _ := config.DefaultIpAddresses()
//
//     i := info{
//         Executable:       executable,
//         WorkDir:          cmd.workdir,
//         LogDir:           cmd.logdir,
//         BindAddress:      &bind,
//         BroadcastAddress: &broadcast,
//     }
//
//     if err := cmd.register(&i); err != nil {
//         return err
//     }
//
//     if err := cmd.mkdirs(&i); err != nil {
//         return err
//     }
//
//     if err := cmd.conf(&i); err != nil {
//         return err
//     }
//
//     if err := cmd.genkeys(&i); err != nil {
//         return err
//     }
//
//     fmt.Printf("   ... %s registered as a Windows system service\n", SERVICE)
//     fmt.Println()
//     fmt.Println("   The service will start automatically on the next system restart. Start it manually from the")
//     fmt.Println("   'Services' application or from the command line by executing the following command:")
//     fmt.Println()
//     fmt.Printf("     > net start %s\n", SERVICE)
//     fmt.Printf("     > sc query %s\n", SERVICE)
//     fmt.Println()
//     fmt.Println("   Please replace the default RSA keys for event and system messages:")
//     fmt.Printf("     - %s\n", filepath.Join(filepath.Dir(cmd.config), "mqtt", "rsa", "encryption", "event.pub"))
//     fmt.Printf("     - %s\n", filepath.Join(filepath.Dir(cmd.config), "mqtt", "rsa", "encryption", "system.pub"))
//     fmt.Println()
//
//     return nil
// }
//
// func (cmd *Daemonize) register(i *info) error {
//     config := mgr.Config{
//         DisplayName:      cmd.name,
//         Description:      cmd.description,
//         StartType:        mgr.StartAutomatic,
//         DelayedAutoStart: true,
//     }
//
//     m, err := mgr.Connect()
//     if err != nil {
//         return err
//     }
//
//     defer m.Disconnect()
//
//     s, err := m.OpenService(cmd.name)
//     if err == nil {
//         s.Close()
//         return fmt.Errorf("service %s already exists", cmd.Name)
//     }
//
//     s, err = m.CreateService(cmd.name, i.Executable, config, "is", "auto-started")
//     if err != nil {
//         return err
//     }
//
//     defer s.Close()
//
//     err = eventlog.InstallAsEventCreate(cmd.name, eventlog.Error|eventlog.Warning|eventlog.Info)
//     if err != nil {
//         s.Delete()
//         return fmt.Errorf("InstallAsEventCreate() failed: %v", err)
//     }
//
//     return nil
// }
//
// func (cmd *Daemonize) mkdirs(i *info) error {
//     directories := []string{
//         i.WorkDir,
//         i.LogDir,
//         filepath.Join(i.WorkDir, "mqtt"),
//         filepath.Join(i.WorkDir, "mqtt", "rsa"),
//         filepath.Join(i.WorkDir, "mqtt", "rsa", "encryption"),
//         filepath.Join(i.WorkDir, "mqtt", "rsa", "signing"),
//     }
//
//     for _, dir := range directories {
//         fmt.Printf("   ... creating '%s'\n", dir)
//
//         if err := os.MkdirAll(dir, 0770); err != nil {
//             return err
//         }
//     }
//
//     return nil
// }
//
// func (cmd *Daemonize) conf(i *info) error {
//     path := cmd.config
//
//     fmt.Printf("   ... creating '%s'\n", path)
//
//     // initialise config from existing uhppoted.conf
//     cfg := config.NewConfig()
//     if f, err := os.Open(path); err != nil {
//         if !os.IsNotExist(err) {
//             return err
//         }
//     } else {
//         err := cfg.Read(f)
//         f.Close()
//         if err != nil {
//             return err
//         }
//     }
//
//     // generate HMAC and RSA keys
//     if cfg.MQTT.HMAC.Key == "" {
//         hmac, err := hmac()
//         if err != nil {
//             return err
//         }
//
//         cfg.MQTT.HMAC.Key = hmac
//     }
//
//     // replace line endings
//     var b strings.Builder
//
//     err := cfg.Write(&b)
//
//     // write back config with any updated information
//     f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
//     if err != nil {
//         return err
//     }
//
//     defer f.Close()
//
//     replacer := strings.NewReplacer(
//         "\r\n", "\r\n",
//         "\r", "\r\n",
//         "\n", "\r\n",
//     )
//
//     if _, err = f.Write([]byte(replacer.Replace(b.String()))); err != nil {
//         return err
//     }
//
//     return nil
// }
//
// func (cmd *Daemonize) genkeys(i *info) error {
//     return genkeys(filepath.Dir(cmd.config), cmd.hotp)
// }
