package commands

import (
	"bufio"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/uhppoted/uhppoted-lib/config"
)

type usergroup string

type info struct {
	Description   string
	Documentation string
	Executable    string
	PID           string
	User          string
	Group         string
	Uid           int
	Gid           int
	LogFiles      []string
}

const serviceTemplate = `[Unit]
Description={{.Description}}
Documentation={{.Documentation}}
After=syslog.target network.target

[Service]
Type=simple
ExecStart={{.Executable}}
PIDFile={{.PID}}
User={{.User}}
Group={{.Group}}

[Install]
WantedBy=multi-user.target
`

const logRotateTemplate = `{{range .LogFiles}}{{.}} {
    daily
    rotate 30
    compress
        compresscmd /bin/bzip2
        compressext .bz2
        dateext
    missingok
    notifempty
    su uhppoted uhppoted
    postrotate
       /usr/bin/killall -HUP uhppoted-httpd
    endscript
}
{{end}}
`

var DAEMONIZE = Daemonize{
	usergroup: "uhppoted:uhppoted",
	workdir:   "/var/uhppoted/httpd",
	logdir:    "/var/log/uhppoted",
	config:    "/etc/uhppoted/uhppoted.conf",
	html:      "/etc/uhppoted/httpd/html",
	etc:       "/etc/uhppoted/httpd",
}

var replacer *strings.Replacer

type Daemonize struct {
	usergroup usergroup
	workdir   string
	logdir    string
	config    string
	html      string
	etc       string
}

func (cmd *Daemonize) Name() string {
	return "daemonize"
}

func (cmd *Daemonize) FlagSet() *flag.FlagSet {
	flagset := flag.NewFlagSet("daemonize", flag.ExitOnError)
	flagset.Var(&cmd.usergroup, "user", "user:group for uhppoted service")

	return flagset
}

func (cmd *Daemonize) Description() string {
	return fmt.Sprintf("Daemonizes %s as a service/daemon", SERVICE)
}

func (cmd *Daemonize) Usage() string {
	return "daemonize [--user <user:group>]"
}

func (cmd *Daemonize) Help() {
	fmt.Println()
	fmt.Println("  Usage: uhppoted daemonize [--user <user:group>]")
	fmt.Println()
	fmt.Printf("    Registers %s as a systemd service/daemon that runs on startup.\n", SERVICE)
	fmt.Println("      Defaults to the user:group uhppoted:uhppoted unless otherwise specified")
	fmt.Println("      with the --user option")
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
	executable, err := os.Executable()
	if err != nil {
		return err
	}

	uid, gid, err := getUserGroup(string(cmd.usergroup))
	if err != nil {
		fmt.Println()
		fmt.Printf("     **** PLEASE CREATE uid:gid %v (OR SPECIFY A DIFFERENT uid:gid WITH the --user OPTION) ***\n", cmd.usergroup)
		fmt.Println()
		return err
	}

	username := "uhppoted"
	if u, err := user.LookupId(fmt.Sprintf("%v", uid)); err == nil {
		username = u.Username
	}

	fmt.Println()
	fmt.Println("   ... daemonizing")

	bind, _, _ := config.DefaultIpAddresses()

	i := info{
		Description:   "UHPPOTE UTO311-L0x access card controllers HTTPD service/daemon ",
		Documentation: "https://github.com/uhppoted/uhppoted-httpd",
		Executable:    executable,
		PID:           fmt.Sprintf("/var/uhppoted/%s.pid", SERVICE),
		User:          "uhppoted",
		Group:         "uhppoted",
		Uid:           uid,
		Gid:           gid,
		LogFiles: []string{
			fmt.Sprintf("/var/log/uhppoted/%s.log", SERVICE),
			"/var/uhppoted/httpd/audit/audit.log",
		},
	}

	chown := func(path string, info fs.DirEntry, err error) error {
		if err == nil {
			err = os.Chown(path, uid, gid)
		}
		return err
	}

	if err := cmd.systemd(&i); err != nil {
		return err
	}

	if err := cmd.mkdirs(&i); err != nil {
		return err
	}

	if err := cmd.logrotate(&i); err != nil {
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
	} else if err = os.Chown(cmd.config, uid, gid); err != nil {
		return err
	}

	if err := cmd.users(i); err != nil {
		return err
	}

	if err := cmd.sysinit(i); err != nil {
		return err
	}

	if _, err := cmd.genTLSkeys(i); err != nil {
		return err
	}

	if err := filepath.WalkDir(cmd.etc, chown); err != nil {
		return err
	}

	if err := filepath.WalkDir(cmd.workdir, chown); err != nil {
		return err
	}

	fmt.Printf("   ... %s registered as a systemd service\n", SERVICE)
	fmt.Println()
	fmt.Println("   The daemon will start automatically on the next system restart - to start it manually, execute the following command:")
	fmt.Println()
	fmt.Printf("     > sudo systemctl start  %s\n", SERVICE)
	fmt.Printf("     > sudo systemctl status %s\n", SERVICE)
	fmt.Println()
	fmt.Println("   The firewall may need additional rules to allow UDP broadcast e.g. for UFW:")
	fmt.Println()
	fmt.Printf("     > sudo ufw allow from %s to any port 60000 proto udp\n", bind.IP)
	fmt.Println()
	fmt.Println("   The firewall may also need additional rules to allow external access to the HTTP server e.g. for UFW:")
	fmt.Println()
	fmt.Printf("     > sudo ufw allow from %s to any port 8080 proto tcp\n", bind.IP)
	fmt.Printf("     > sudo ufw allow from %s to any port 8443 proto tcp\n", bind.IP)
	fmt.Println()
	fmt.Printf("   The installation can be verified by running the %v service in 'console' mode:\n", SERVICE)
	fmt.Println()
	fmt.Printf("     > sudo su %v\n", username)
	fmt.Printf("     > ./uhppoted-httpd --debug --console\n")
	fmt.Println()
	fmt.Println()

	return nil
}

func (cmd *Daemonize) systemd(i *info) error {
	service := fmt.Sprintf("%s.service", SERVICE)
	path := filepath.Join("/etc/systemd/system", service)
	t := template.Must(template.New(service).Parse(serviceTemplate))

	fmt.Printf("   ... creating '%s'\n", path)
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	defer f.Close()

	return t.Execute(f, i)
}

func (cmd *Daemonize) mkdirs(i *info) error {
	directories := []string{
		"/var/uhppoted",
		"/var/uhppoted/httpd",
		"/var/log/uhppoted",
		"/etc/uhppoted",
		"/etc/uhppoted/httpd",
	}

	for _, dir := range directories {
		fmt.Printf("   ... creating '%s'\n", dir)

		if err := os.MkdirAll(dir, 0770); err != nil {
			return err
		}

		if err := os.Chown(dir, i.Uid, i.Gid); err != nil {
			return err
		}
	}

	return nil
}

func (cmd *Daemonize) logrotate(i *info) error {
	path := filepath.Join("/etc/logrotate.d", SERVICE)
	t := template.Must(template.New(fmt.Sprintf("%s.logrotate", SERVICE)).Parse(logRotateTemplate))

	fmt.Printf("   ... creating '%s'\n", path)
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	defer f.Close()

	return t.Execute(f, i)
}

// usergroup::flag.Value
func getUserGroup(s string) (int, int, error) {
	match := regexp.MustCompile(`(\w+?):(\w+)`).FindStringSubmatch(s)
	if match == nil {
		return 0, 0, fmt.Errorf("Invalid user:group '%s'", s)
	}

	u, err := user.Lookup(match[1])
	if err != nil {
		return 0, 0, err
	}

	g, err := user.LookupGroup(match[2])
	if err != nil {
		return 0, 0, err
	}

	uid, err := strconv.Atoi(u.Uid)
	if err != nil {
		return 0, 0, err
	}

	gid, err := strconv.Atoi(g.Gid)
	if err != nil {
		return 0, 0, err
	}

	return uid, gid, nil
}

func (f *usergroup) String() string {
	if f == nil {
		return "uhppoted:uhppoted"
	}

	return string(*f)
}

func (f *usergroup) Set(s string) error {
	_, _, err := getUserGroup(s)
	if err != nil {
		return err
	}

	*f = usergroup(strings.TrimSpace(s))

	return nil
}
