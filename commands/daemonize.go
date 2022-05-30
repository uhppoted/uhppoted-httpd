package commands

import (
	"bufio"
	"bytes"
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
      "path": "^/sys/overview.html$",
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
      "path": "^/sys/overview.html$",
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
    },
    {
      "path": "^/synchronize/ACL$",
      "authorised": "^(admin)$"
    },
    {
      "path": "^/synchronize/datetime$",
      "authorised": "^(admin)$"
    }
  ]
}
`

func (cmd *Daemonize) conf(i info, unpacked bool, grules bool) error {
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

	return toTextFile(b.String(), path)
}

func (cmd *Daemonize) unpack(i info) (bool, error) {
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

		if path == "html.go" {
			return nil
		}

		file := filepath.Join(root, path)
		if _, err := os.Stat(file); err != nil && os.IsNotExist(err) {
			src, err := html.HTML.Open(path)
			if err != nil {
				return err
			}

			defer src.Close()

			var b bytes.Buffer
			if _, err := io.Copy(&b, src); err != nil {
				return err
			}

			switch filepath.Ext(file) {
			case ".html", ".json", ".js", ".css":
				if err := toTextFile(b.String(), file); err != nil {
					return err
				}

			default:
				if err := toBinaryFile(b.Bytes(), file); err != nil {
					return err
				}
			}
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

func (cmd *Daemonize) grules(i info) (bool, error) {
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

			return toTextFile(b.String(), file)
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
	fmt.Println()
	if err != nil || strings.ToLower(strings.TrimSpace(text)) != "yes" {
		return nil
	}

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
		OID      string `json:"OID"`
		UID      string `json:"uid,omitempty"`
		Role     string `json:"role,omitempty"`
		Salt     string `json:"salt"`
		Password string `json:"password"`
	}{
		OID:      "0.8.1",
		UID:      "admin",
		Role:     "admin",
		Salt:     hex.EncodeToString(salt[:]),
		Password: fmt.Sprintf("%0x", h.Sum(nil)),
	}

	users.Users = append(users.Users, admin)

	// ... write default 'admin' user to users.json
	if bytes, err := json.MarshalIndent(users, "  ", "  "); err != nil {
		return err
	} else if err := toTextFile(string(bytes), file); err != nil {
		return err
	}

	fmt.Printf("   ... created default 'admin' user, password:%v\n", string(password))
	fmt.Println()

	return nil
}

func (cmd *Daemonize) sysinit(i info) error {
	// ... create empty system files
	folder := filepath.Join(cmd.workdir, "system")

	fmt.Printf("   ... creating folder '%v'\n", folder)
	if err := os.MkdirAll(folder, 0744); err != nil {
		return err
	}

	files := []struct {
		file    string
		content string
	}{
		{"interfaces.json", interfaces},
		{"controllers.json", controllers},
		{"doors.json", doors},
		{"cards.json", cards},
		{"groups.json", groups},
		{"events.json", events},
		{"logs.json", logs},
	}

	for _, v := range files {
		file := filepath.Join(folder, v.file)

		if _, err := os.Stat(file); err != nil {
			if !os.IsNotExist(err) {
				return err
			}

			fmt.Printf("   ... creating default %v\n", file)
			if err := toTextFile(v.content, file); err != nil {
				return err
			}
		}
	}

	// ... create default auth.json file
	file := filepath.Join(cmd.etc, "auth.json")
	if _, err := os.Stat(file); err != nil {
		if !os.IsNotExist(err) {
			return err
		}

		fmt.Println("   ... creating default 'auth.json'")
		if err := toTextFile(default_auth, file); err != nil {
			return err
		}
	}

	// ... create default acl.grl file
	file = filepath.Join(cmd.etc, "acl.grl")
	if _, err := os.Stat(file); err != nil {
		if !os.IsNotExist(err) {
			return err
		}

		fmt.Println("   ... creating default 'acl.grl'")
		if err := toTextFile(acl, file); err != nil {
			return err
		}
	}

	return nil
}

func (cmd *Daemonize) genTLSkeys(i info) (bool, error) {
	root := cmd.etc
	r := bufio.NewReader(os.Stdin)

	fmt.Println()
	fmt.Printf("     Do you want to create TLS keys and certificates (yes/no)? ")

	text, err := r.ReadString('\n')
	fmt.Println()
	if err != nil || strings.TrimSpace(text) != "yes" {
		return false, nil
	}

	keys, err := genkeys()
	if err != nil {
		return false, err
	} else if keys == nil {
		return false, fmt.Errorf("Invalid TLS key set (%v)", keys)
	}

	list := []struct {
		item interface{}
		file string
	}{
		{keys.CA.privateKey, "ca.key"},
		{keys.CA.certificate, "ca.cert"},
		{keys.server.privateKey, "uhppoted.key"},
		{keys.server.certificate, "uhppoted.cert"},
		{keys.client.privateKey, "client.key"},
		{keys.client.certificate, "client.cert"},
	}

	for _, v := range list {
		file := filepath.Join(root, v.file)

		//	if _, err := os.Stat(file); err != nil {
		//		if !os.IsNotExist(err) {
		//			return false, err
		//		} else if err := toTextFile(string(encode(v.item)), file); err != nil {
		//			return false, err
		//		} else {
		//			fmt.Printf("   ... created %v\n", file)
		//		}
		//	}

		if err := toTextFile(string(encode(v.item)), file); err != nil {
			return false, err
		} else {
			fmt.Printf("   ... created %v\n", file)
		}
	}

	fmt.Println()
	fmt.Println("   ** PLEASE MOVE THE ca.key FILE TO A SECURE LOCATION")
	fmt.Println()
	fmt.Println("   The supplied client.key file can be installed in a browser to support mutual TLS authentication.")
	fmt.Println("   It is provided merely as an example and both the client key and certificate should be removed")
	fmt.Println("   and replaced by your own keys and certificates.")
	fmt.Println()
	fmt.Println("   The client.key file is in PEM format - to convert it to a PKCS12 file for importing into Firefox")
	fmt.Println("   execute the following command:")
	fmt.Println()
	fmt.Println("   openssl pkcs12 -export -in client.cert -inkey client.key -certfile ca.cert -out client.p12")
	fmt.Println()
	fmt.Println("   ** NB: The generated TLS keys and certificates are for TEST USE ONLY and should be replaced with")
	fmt.Println("          your own CA certificate and server and client keys and certificates for production use.")
	fmt.Println()

	return true, nil
}

func toTextFile(buffer string, path string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0660)
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

func toBinaryFile(buffer []byte, path string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0660)
	if err != nil {
		return err
	}

	defer f.Close()

	if _, err = f.Write(buffer); err != nil {
		return err
	}

	return nil
}
