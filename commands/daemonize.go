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

	"github.com/uhppoted/uhppoted-httpd/httpd/html"
)

func (cmd *Daemonize) conf(i *info, unpacked bool) error {
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
		cfg.HTTPD.HTML = cmd.html
	}

	// ... write back with added HTTPD config
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	defer f.Close()

	return cfg.Write(f)
}

func (cmd *Daemonize) unpack(i *info) (bool, error) {
	root := cmd.html
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
	fmt.Printf("     ... copying files to %v\n", root)
	if err := fs.WalkDir(html.HTML, ".", cp); err != nil {
		return false, err
	}

	fmt.Println()

	return true, nil
}

func (cmd *Daemonize) users(i *info) error {
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
	} else if err := os.WriteFile(file, bytes, 0660); err != nil {
		return err
	}

	fmt.Printf("   ... created default 'admin' user, password:%v\n", string(password))

	return nil
}
