package commands

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/uhppoted/uhppoted-lib/config"

	"github.com/uhppoted/uhppoted-httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/httpd/html"
)

const interfaces = `{
  "interfaces": [
    {
      "OID": "0.1.1",
      "name": "LAN",
      "bind-address": "0.0.0.0",
      "broadcast-address": "255.255.255.255",
      "listen-address": "0.0.0.0:60001"
    }
  ]
}
`

const controllers = `{ "controllers": [] }`
const doors = `{ "doors": [] }`
const cards = `{ "cards": [] }`
const groups = `{ "groups": [] }`
const events = `{ "events": [] }`
const logs = `{ "logs": [] }`
const acl = ``

const default_auth = `
{
  "users": {},
  "resources": [
    {
      "path": "^/index.html$",
      "authorised": ".*"
    },
    {
      "path": "^/favicon.ico$",
      "authorised": ".*"
    },
    {
      "path": "^/sys/login.html$",
      "authorised": ".*"
    },
    {
      "path": "^/sys/unauthorized.html$",
      "authorised": ".*"
    },
    {
      "path": "^/sys/controllers.html$",
      "authorised": "^(admin)$"
    },
    {
      "path": "^/sys/cards.html$",
      "authorised": "^(admin|user)$"
    },
    {
      "path": "^/sys/doors.html$",
      "authorised": "^(admin)$"
    },
    {
      "path": "^/sys/groups.html$",
      "authorised": "^(admin)$"
    },
    {
      "path": "^/sys/events.html$",
      "authorised": "^(admin)$"
    },
    {
      "path": "^/sys/logs.html$",
      "authorised": "^(admin)$"
    },
    {
      "path": "^/sys/users.html$",
      "authorised": "^(admin)$"
    },
    {
      "path": "^/sys/password.html$",
      "authorised": ".*"
    },
    {
      "path": "^/other.html$",
      "authorised": ".*"
    },
    {
      "path": "^/password$",
      "authorised": ".*"
    },
    {
      "path": "^/interfaces$",
      "authorised": "^(admin)$"
    },
    {
      "path": "^/controllers$",
      "authorised": "^(admin)$"
    },
    {
      "path": "^/doors$",
      "authorised": "^(admin)$"
    },
    {
      "path": "^/cards$",
      "authorised": "^(admin|user)$"
    },
    {
      "path": "^/groups$",
      "authorised": "^(admin|user)$"
    },
    {
      "path": "^/events$",
      "authorised": "^(admin)$"
    },
    {
      "path": "^/logs$",
      "authorised": "^(admin)$"
    },
    {
      "path": "^/users$",
      "authorised": "^(admin)$"
    }
  ]
}
`

func (cmd *Daemonize) conf(i *info, unpacked bool, grules bool) error {
	path := cmd.config

	fmt.Printf("   ... creating '%s'\n", path)

	// ... get config from existing uhppoted.conf
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
		cfg.HTTPD.HTML = filepath.Join(cmd.etc, "html")
	}

	// ... update rules files if grules files unpacked
	if grules {
		cfg.HTTPD.DB.Rules.Interfaces = filepath.Join(cmd.etc, "grules", "interfaces.grl")
		cfg.HTTPD.DB.Rules.Controllers = filepath.Join(cmd.etc, "grules", "controllers.grl")
		cfg.HTTPD.DB.Rules.Doors = filepath.Join(cmd.etc, "grules", "doors.grl")
		cfg.HTTPD.DB.Rules.Cards = filepath.Join(cmd.etc, "grules", "cards.grl")
		cfg.HTTPD.DB.Rules.Groups = filepath.Join(cmd.etc, "grules", "groups.grl")
		cfg.HTTPD.DB.Rules.Events = filepath.Join(cmd.etc, "grules", "events.grl")
		cfg.HTTPD.DB.Rules.Logs = filepath.Join(cmd.etc, "grules", "logs.grl")
		cfg.HTTPD.DB.Rules.Users = filepath.Join(cmd.etc, "grules", "users.grl")
	}

	// ... write back with added HTTPD config
	var b strings.Builder
	if err := cfg.Write(&b); err != nil {
		return err
	}

	return write(b.String(), path)
}

func (cmd *Daemonize) unpack(i *info) (bool, error) {
	root := filepath.Join(cmd.etc, "html")
	r := bufio.NewReader(os.Stdin)

	fmt.Println()
	fmt.Printf("     Do you want to unpack the HTML files into %v (yes/no)? ", root)

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

		file := filepath.Join(root, path)
		if _, err := os.Stat(file); err != nil && os.IsNotExist(err) {
			src, err := html.HTML.Open(path)
			if err != nil {
				return err
			}

			defer src.Close()

			var b strings.Builder
			if _, err := io.Copy(&b, src); err != nil {
				return err
			}

			return write(b.String(), file)
		}

		return nil
	}

	fmt.Println()
	fmt.Printf("     ... copying files to %v\n", root)
	if err := fs.WalkDir(html.HTML, ".", cp); err != nil {
		return false, err
	}

	fmt.Println()

	return true, nil
}

func (cmd *Daemonize) grules(i *info) (bool, error) {
	root := cmd.etc
	r := bufio.NewReader(os.Stdin)

	fmt.Println()
	fmt.Printf("     Do you want to unpack the GRULES files into %v (yes/no)? ", filepath.Join(root, "grules"))

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

	if err := fs.WalkDir(auth.GRULES, ".", mkdir); err != nil {
		return false, err
	}

	cp := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		} else if d.IsDir() {
			return nil
		}

		file := filepath.Join(root, path)
		if _, err := os.Stat(file); err != nil && os.IsNotExist(err) {
			src, err := auth.GRULES.Open(path)
			if err != nil {
				return err
			}

			defer src.Close()

			var b strings.Builder
			if _, err := io.Copy(&b, src); err != nil {
				return err
			}

			return write(b.String(), file)
		}

		return nil
	}

	fmt.Printf("     ... copying GRULES files to %v\n", filepath.Join(root, "grules"))
	if err := fs.WalkDir(auth.GRULES, ".", cp); err != nil {
		return false, err
	}

	fmt.Println()

	return true, nil
}

func (cmd *Daemonize) users(i info) error {
	dir := filepath.Join(cmd.workdir, "system")

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

	// ... create initial 'admin' user?
	stdin := bufio.NewReader(os.Stdin)

	fmt.Println()
	fmt.Printf("     Do you want to create a default 'admin' user (yes/no)? ")

	text, err := stdin.ReadString('\n')
	if err != nil || strings.ToLower(strings.TrimSpace(text)) == "no" {
		fmt.Println()
		return nil
	}

	fmt.Println()
	fmt.Println("   ... creating default 'admin' user")

	// ... generate password and salt
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
	} else if err := write(string(bytes), file); err != nil {
		return err
	}

	fmt.Printf("   ... created default 'admin' user, password:%v\n", string(password))
	fmt.Println()

	return nil
}

func (cmd *Daemonize) sysinit(i info) error {
	// ... create empty system files
	folder := filepath.Join(cmd.etc, "system")

	fmt.Printf("   ... creating folder '%v'\n", folder)
	if err := os.MkdirAll(folder, 0744); err != nil {
		return err
	}

	files := []struct {
		file    string
		content []byte
	}{
		{"interfaces.json", []byte(interfaces)},
		{"controllers.json", []byte(controllers)},
		{"doors.json", []byte(doors)},
		{"cards.json", []byte(cards)},
		{"groups.json", []byte(groups)},
		{"events.json", []byte(events)},
		{"logs.json", []byte(logs)},
	}

	for _, v := range files {
		file := filepath.Join(folder, v.file)

		if _, err := os.Stat(file); err != nil {
			if os.IsNotExist(err) {
				fmt.Println("   ... creating default 'interfaces.json'")
				if err := os.WriteFile(file, v.content, 0660); err != nil {
					return err
				}
			} else {
				return err
			}
		}
	}

	// ... create default auth.json file
	file := filepath.Join(cmd.etc, "auth.json")
	if _, err := os.Stat(file); err != nil {
		if os.IsNotExist(err) {
			fmt.Println("   ... creating default 'auth.json'")
			if err := write(default_auth, file); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// ... create default acl.grl file
	file = filepath.Join(cmd.etc, "acl.grl")
	if _, err := os.Stat(file); err != nil {
		if os.IsNotExist(err) {
			fmt.Println("   ... creating default 'acl.grl'")
			if err := write(acl, file); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	return nil
}

func write(buffer string, path string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	defer f.Close()

	s := buffer
	if replacer != nil {
		s = replacer.Replace(buffer)
	}

	if _, err = f.WriteString(s); err != nil {
		return err
	}

	return nil
}
