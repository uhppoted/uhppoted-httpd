package commands

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/uhppoted/uhppoted-lib/config"
	xpath "github.com/uhppoted/uhppoted-lib/encoding/plist"

	"github.com/uhppoted/uhppoted-httpd/httpd/html"
)

type info struct {
	Label      string
	Executable string
	WorkDir    string
	HTML       string
	StdLogFile string
	ErrLogFile string
}

type plist struct {
	Label             string
	Program           string
	WorkingDirectory  string
	ProgramArguments  []string
	KeepAlive         bool
	RunAtLoad         bool
	StandardOutPath   string
	StandardErrorPath string
}

const newsyslog = `#logfilename                                       [owner:group]  mode  count  size   when  flags [/pid_file]  [sig_num]
{{range .}}{{.LogFile}}  :              644   30     10000  @T00  J     {{.PID}}
{{end}}`

var DAEMONIZE = Daemonize{
	plist:   fmt.Sprintf("com.github.uhppoted.%s.plist", SERVICE),
	workdir: "/usr/local/var/com.github.uhppoted/httpd",
	logdir:  "/usr/local/var/com.github.uhppoted/logs",
	config:  "/usr/local/etc/com.github.uhppoted/uhppoted.conf",
	html:    "/usr/local/etc/com.github.uhppoted/httpd/html",
}

type Daemonize struct {
	plist   string
	workdir string
	logdir  string
	config  string
	html    string
}

func (cmd *Daemonize) Name() string {
	return "daemonize"
}

func (cmd *Daemonize) FlagSet() *flag.FlagSet {
	return flag.NewFlagSet("daemonize", flag.ExitOnError)
}

func (cmd *Daemonize) Description() string {
	return fmt.Sprintf("Daemonizes %s as a service/daemon", SERVICE)
}

func (cmd *Daemonize) Usage() string {
	return ""
}

func (cmd *Daemonize) Help() {
	fmt.Println()
	fmt.Printf("  Usage: %s daemonize\n", SERVICE)
	fmt.Println()
	fmt.Printf("    Daemonizes %s as a service/daemon that runs on startup\n", SERVICE)
	fmt.Println()

	helpOptions(cmd.FlagSet())
}

func (cmd *Daemonize) Execute(args ...interface{}) error {
	dir := filepath.Dir(cmd.config)
	r := bufio.NewReader(os.Stdin)

	fmt.Println()
	fmt.Printf("     **** PLEASE MAKE SURE YOU HAVE A BACKUP COPY OF ANY CONFIGURATION INFORMATION AND KEYS IN %s ***\n", dir)
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

	i := info{
		Label:      fmt.Sprintf("com.github.uhppoted.%s", SERVICE),
		Executable: executable,
		WorkDir:    cmd.workdir,
		HTML:       cmd.html,
		StdLogFile: filepath.Join(cmd.logdir, fmt.Sprintf("%s.log", SERVICE)),
		ErrLogFile: filepath.Join(cmd.logdir, fmt.Sprintf("%s.err", SERVICE)),
	}

	if err := cmd.launchd(&i); err != nil {
		return err
	}

	if err := cmd.mkdirs(); err != nil {
		return err
	}

	if err := cmd.logrotate(&i); err != nil {
		return err
	}

	if err := cmd.firewall(&i); err != nil {
		return err
	}

	unpacked, err := cmd.unpack(&i)
	if err != nil {
		return err
	}

	if err := cmd.conf(&i, unpacked); err != nil {
		return err
	}

	if err := cmd.users(&i); err != nil {
		return err
	}

	fmt.Printf("   ... %s registered as a LaunchDaemon\n", i.Label)
	fmt.Println()
	fmt.Printf("   The daemon will start automatically on the next system restart - to start it manually, execute the following command:\n")
	fmt.Println()
	fmt.Printf("   sudo launchctl load /Library/LaunchDaemons/com.github.uhppoted.%s.plist\n", SERVICE)
	fmt.Println()
	fmt.Println()

	return nil
}

func (cmd *Daemonize) launchd(i *info) error {
	path := filepath.Join("/Library/LaunchDaemons", cmd.plist)
	_, err := os.Stat(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	pl := plist{
		Label:             i.Label,
		Program:           i.Executable,
		WorkingDirectory:  i.WorkDir,
		ProgramArguments:  []string{},
		KeepAlive:         true,
		RunAtLoad:         true,
		StandardOutPath:   i.StdLogFile,
		StandardErrorPath: i.ErrLogFile,
	}

	if !os.IsNotExist(err) {
		current, err := cmd.parse(path)
		if err != nil {
			return err
		}

		pl.WorkingDirectory = current.WorkingDirectory
		pl.ProgramArguments = current.ProgramArguments
		pl.KeepAlive = current.KeepAlive
		pl.RunAtLoad = current.RunAtLoad
		pl.StandardOutPath = current.StandardOutPath
		pl.StandardErrorPath = current.StandardErrorPath
	}

	return cmd.daemonize(path, pl)
}

func (cmd *Daemonize) parse(path string) (*plist, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	p := plist{}
	decoder := xpath.NewDecoder(f)
	err = decoder.Decode(&p)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (cmd *Daemonize) daemonize(path string, p interface{}) error {
	fmt.Printf("   ... creating '%s'\n", path)
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	defer f.Close()

	encoder := xpath.NewEncoder(f)
	if err = encoder.Encode(p); err != nil {
		return err
	}

	return nil
}

func (cmd *Daemonize) mkdirs() error {
	directories := []string{
		cmd.workdir,
		cmd.logdir,
	}

	for _, dir := range directories {
		fmt.Printf("   ... creating '%s'\n", dir)

		if err := os.MkdirAll(dir, 0644); err != nil {
			return err
		}
	}

	return nil
}

func (cmd *Daemonize) conf(i *info, unpacked bool) error {
	path := cmd.config

	fmt.Printf("   ... creating '%s'\n", path)

	// ... gGet config from existing uhppoted.conf
	cfg := config.NewConfig()
	if f, err := os.Open(path); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		err := cfg.Read(f)
		f.Close()
		if err != nil {
			return err
		}
	}

	// ... update httpd.HTML if unpacked
	if unpacked {
		cfg.HTTPD.HTML = i.HTML
	}

	// ... write back with added HTTPD config
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	defer f.Close()

	return cfg.Write(f)
}

func (cmd *Daemonize) logrotate(i *info) error {
	pid := filepath.Join(cmd.workdir, fmt.Sprintf("%s.pid", SERVICE))
	logfiles := []struct {
		LogFile string
		PID     string
	}{
		{
			LogFile: i.StdLogFile,
			PID:     pid,
		},
		{
			LogFile: i.ErrLogFile,
			PID:     pid,
		},
	}

	t := template.Must(template.New("logrotate.conf").Parse(newsyslog))
	path := filepath.Join("/etc/newsyslog.d", fmt.Sprintf("%s.conf", SERVICE))

	fmt.Printf("   ... creating '%s'\n", path)

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	defer f.Close()

	return t.Execute(f, logfiles)
}

func (cmd *Daemonize) firewall(i *info) error {
	fmt.Println()
	fmt.Printf("   ***\n")
	fmt.Printf("   *** WARNING: adding '%s' to the application firewall and unblocking incoming connections\n", SERVICE)
	fmt.Printf("   ***\n")
	fmt.Println()

	path := i.Executable

	command := exec.Command("/usr/libexec/ApplicationFirewall/socketfilterfw", "--getglobalstate")
	out, err := command.CombinedOutput()
	fmt.Printf("   > %s", out)
	if err != nil {
		return fmt.Errorf("Failed to retrieve application firewall global state (%v)", err)
	}

	if strings.Contains(string(out), "State = 1") {
		command = exec.Command("/usr/libexec/ApplicationFirewall/socketfilterfw", "--setglobalstate", "off")
		out, err = command.CombinedOutput()
		fmt.Printf("   > %s", out)
		if err != nil {
			return fmt.Errorf("Failed to disable the application firewall (%v)", err)
		}

		command = exec.Command("/usr/libexec/ApplicationFirewall/socketfilterfw", "--add", path)
		out, err = command.CombinedOutput()
		fmt.Printf("   > %s", out)
		if err != nil {
			return fmt.Errorf("Failed to add 'uhppoted-rest' to the application firewall (%v)", err)
		}

		command = exec.Command("/usr/libexec/ApplicationFirewall/socketfilterfw", "--unblockapp", path)
		out, err = command.CombinedOutput()
		fmt.Printf("   > %s", out)
		if err != nil {
			return fmt.Errorf("Failed to unblock 'uhppoted-rest' on the application firewall (%v)", err)
		}

		command = exec.Command("/usr/libexec/ApplicationFirewall/socketfilterfw", "--setglobalstate", "on")
		out, err = command.CombinedOutput()
		fmt.Printf("   > %s", out)
		if err != nil {
			return fmt.Errorf("Failed to re-enable the application firewall (%v)", err)
		}

		fmt.Println()
	}

	return nil
}

func (cmd *Daemonize) unpack(i *info) (bool, error) {
	root := i.HTML
	r := bufio.NewReader(os.Stdin)

	fmt.Println()
	fmt.Printf("     Do you want to unpack the HTML files into %v (yes/no)? ", i.HTML)

	text, err := r.ReadString('\n')
	if err != nil || strings.TrimSpace(text) != "yes" {
		fmt.Println()
		return false, nil
	}

	fmt.Println()

	mkdir := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		} else if d.IsDir() {
			folder := filepath.Join(root, path)

			fmt.Printf("     ... creating folder '%v'\n", folder)
			if err := os.MkdirAll(folder, 0744); err != nil {
				return err
			}
		}

		return nil
	}

	if err := fs.WalkDir(html.HTML, ".", mkdir); err != nil {
		return false, err
	}

	cp := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		} else if d.IsDir() {
			return nil
		}

		src, err := html.HTML.Open(path)
		if err != nil {
			return err
		}

		defer src.Close()

		dest, err := os.Create(filepath.Join(root, path))
		if err != nil {
			return err
		}

		defer dest.Close()

		if _, err := io.Copy(dest, src); err != nil {
			return err
		}

		return nil
	}

	fmt.Println()
	fmt.Printf("     ... copying files to %v\n", i.HTML)
	if err := fs.WalkDir(html.HTML, ".", cp); err != nil {
		return false, err
	}

	fmt.Println()

	return true, nil
}

func (cmd *Daemonize) users(i *info) error {
	dir := filepath.Join(i.WorkDir, "system")

	fmt.Printf("   ... creating folder '%v'\n", dir)
	if err := os.MkdirAll(dir, 0744); err != nil {
		return err
	}

	file := filepath.Join(dir, "users.json")
	users := struct {
		Users []interface{} `json:"users"`
	}{}

	info, err := os.Stat(file)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else if !info.IsDir() {
		bytes, err := os.ReadFile(file)
		if err != nil {
			return err
		}

		if err := json.Unmarshal(bytes, &users); err != nil {
			return err
		}
	}

	if len(users.Users) > 0 {
		return nil
	}

	// ... create initial 'admin' user

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	letters := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	password := make([]byte, 16)
	salt := make([]byte, 16)

	for i := range password {
		password[i] = letters[r.Intn(len(letters))]
	}

	for i := range salt {
		salt[i] = byte(r.Intn(256))
	}

	h := sha256.New()
	h.Write(salt)
	h.Write([]byte(password))

	admin := struct {
		UID      string `json:"uid,omitempty"`
		Role     string `json:"role,omitempty"`
		Salt     string `json:"salt"`
		Password string `json:"password"`
	}{
		UID:      "admin",
		Role:     "admin",
		Salt:     hex.EncodeToString(salt[:]),
		Password: fmt.Sprintf("%0x", h.Sum(nil)),
	}

	users.Users = append(users.Users, admin)

	// ... write default 'admin' user to users.json
	if bytes, err := json.MarshalIndent(users, "  ", "  "); err != nil {
		return err
	} else if err := os.WriteFile(file, bytes, 0660); err != nil {
		return err
	}

	fmt.Printf("   ... created default 'admin' user, password:%v\n", string(password))

	return nil
}
