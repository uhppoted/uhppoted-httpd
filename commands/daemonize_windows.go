package commands

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"

	"github.com/uhppoted/uhppote-core/types"
	"github.com/uhppoted/uhppoted-lib/config"
)

var DAEMONIZE = Daemonize{
	name:        SERVICE,
	description: "UHPPOTE UTO311-L0x access card controllers HTTP service",
	workdir:     filepath.Join(workdir(), "httpd"),
	logdir:      filepath.Join(workdir(), "logs"),
	config:      filepath.Join(workdir(), "uhppoted.conf"),
	etc:         filepath.Join(workdir(), "httpd"),
}

type info struct {
	Executable       string
	WorkDir          string
	HTML             string
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
	etc         string
}

var replacer = strings.NewReplacer(
	"\r\n", "\r\n",
	"\r", "\r\n",
	"\n", "\r\n",
)

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
	dir := filepath.Dir(cmd.config)
	r := bufio.NewReader(os.Stdin)

	fmt.Println()
	fmt.Printf("     **** PLEASE MAKE SURE YOU HAVE A BACKUP COPY OF THE CONFIGURATION INFORMATION AND KEYS IN %s ***\n", dir)
	fmt.Println()
	fmt.Printf("     Enter 'yes' to continue with the installation: ")

	text, err := r.ReadString('\n')
	if err != nil || strings.TrimSpace(text) != "yes" {
		fmt.Println()
		fmt.Printf("     -- installation cancelled --")
		fmt.Println()
		return nil
	}

	return cmd.execute()
}

func (cmd *Daemonize) execute() error {
	fmt.Println()
	fmt.Println("   ... daemonizing")

	executable, err := os.Executable()
	if err != nil {
		return err
	}

	bind, broadcast, _ := config.DefaultIpAddresses()

	i := info{
		Executable:       executable,
		WorkDir:          cmd.workdir,
		LogDir:           cmd.logdir,
		BindAddress:      &bind,
		BroadcastAddress: &broadcast,
	}

	if err := cmd.register(&i); err != nil {
		return err
	}

	if err := cmd.mkdirs(&i); err != nil {
		return err
	}

	unpacked, err := cmd.unpack(i)
	if err != nil {
		return err
	}

	grules, err := cmd.grules(i)
	if err != nil {
		return err
	}

	if err := cmd.conf(i, unpacked, grules); err != nil {
		return err
	}

	admin, pwd, err := cmd.users(i)
	if err != nil {
		return err
	}

	if err := cmd.sysinit(i); err != nil {
		return err
	}

	if _, err := cmd.genTLSkeys(i); err != nil {
		return err
	}

	fmt.Printf("   ... %s registered as a Windows system service\n", SERVICE)
	fmt.Println()
	fmt.Println("   The service will start automatically on the next system restart. Start it manually from the")
	fmt.Println("   'Services' application or from the command line by executing the following command:")
	fmt.Println()
	fmt.Printf("     > net start %s\n", SERVICE)
	fmt.Printf("     > sc query %s\n", SERVICE)
	fmt.Println()

	if admin != "" {
		fmt.Println()
		fmt.Printf("   *** THE %v USER PASSWORD IS %v ***\n", admin, pwd)
		fmt.Println()
	}

	fmt.Println()

	return nil
}

func (cmd *Daemonize) register(i *info) error {
	config := mgr.Config{
		DisplayName:      cmd.name,
		Description:      cmd.description,
		StartType:        mgr.StartAutomatic,
		DelayedAutoStart: true,
	}

	m, err := mgr.Connect()
	if err != nil {
		return err
	}

	defer m.Disconnect()

	s, err := m.OpenService(cmd.name)
	if err == nil {
		s.Close()
		return fmt.Errorf("service %v already exists", cmd.Name)
	}

	s, err = m.CreateService(cmd.name, i.Executable, config, "is", "auto-started")
	if err != nil {
		return err
	}

	defer s.Close()

	err = eventlog.InstallAsEventCreate(cmd.name, eventlog.Error|eventlog.Warning|eventlog.Info)
	if err != nil {
		s.Delete()
		return fmt.Errorf("InstallAsEventCreate() failed: %v", err)
	}

	return nil
}

func (cmd *Daemonize) mkdirs(i *info) error {
	directories := []string{
		i.WorkDir,
		i.LogDir,
	}

	for _, dir := range directories {
		fmt.Printf("   ... creating '%s'\n", dir)

		if err := os.MkdirAll(dir, 0770); err != nil {
			return err
		}
	}

	return nil
}
