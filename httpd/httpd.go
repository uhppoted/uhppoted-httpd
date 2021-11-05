package httpd

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/uhppoted/uhppoted-httpd/httpd/auth"
)

type HTTPD struct {
	Dir                      string
	AuthProvider             auth.IAuth
	HTTPEnabled              bool
	HTTPSEnabled             bool
	CACertificate            string
	TLSCertificate           string
	TLSKey                   string
	RequireClientCertificate bool
	RequestTimeout           time.Duration
	DB                       struct {
		GRules struct {
			System string
			Cards  string
			Doors  string
			Groups string
			ACL    string
		}
	}
}

type dispatcher struct {
	root  string
	fs    http.Handler
	auth  auth.IAuth
	grule struct {
		system string
		cards  string
		doors  string
		groups string
	}
	timeout time.Duration
}

const (
	SettingsCookie = "uhppoted-settings"
)

func (h *HTTPD) Run() {
	fs := httpdFileSystem{
		FileSystem: http.Dir(h.Dir),
	}

	d := dispatcher{
		root: h.Dir,
		fs:   http.FileServer(fs),
		auth: h.AuthProvider,
		grule: struct {
			system string
			cards  string
			doors  string
			groups string
		}{
			system: h.DB.GRules.System,
			cards:  h.DB.GRules.Cards,
			doors:  h.DB.GRules.Doors,
			groups: h.DB.GRules.Groups,
		},
		timeout: h.RequestTimeout,
	}

	var srv *http.Server
	var srvs *http.Server

	if h.HTTPEnabled {
		srv = &http.Server{
			Addr: ":8080",
		}
	}

	if h.HTTPSEnabled {
		ca, err := ioutil.ReadFile(h.CACertificate)
		if err != nil {
			log.Fatal(fmt.Errorf("Error reading CA certificate file '%s' (%v)", h.CACertificate, err))
		}

		certificates := x509.NewCertPool()
		if !certificates.AppendCertsFromPEM(ca) {
			log.Fatal("Unable failed to parse CA certificate")
		}

		tlsConfig := tls.Config{
			ClientCAs: certificates,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			},
			PreferServerCipherSuites: true,
			MinVersion:               tls.VersionTLS12,
		}

		if h.RequireClientCertificate {
			tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		} else {
			tlsConfig.ClientAuth = tls.VerifyClientCertIfGiven
		}

		tlsConfig.BuildNameToCertificate()

		srvs = &http.Server{
			Addr:      ":8443",
			TLSConfig: &tlsConfig,
		}
	}

	shutdown := make(chan struct{})

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		if srv != nil {
			if err := srv.Shutdown(context.Background()); err != nil {
				log.Printf("WARN  HTTP  server shutdown error: %v", err)
			}
		}

		if srvs != nil {
			if err := srvs.Shutdown(context.Background()); err != nil {
				log.Printf("WARN  HTTPS server shutdown error: %v", err)
			}
		}

		close(shutdown)
	}()

	http.Handle("/", &d)

	if srv != nil {
		go func() {
			log.Printf("INFO  HTTP  server starting on port %v", srv.Addr)
			if err := srv.ListenAndServe(); err != http.ErrServerClosed {
				log.Fatalf("ERROR: %v", err)
			}
		}()
	}

	if srvs != nil {
		go func() {
			log.Printf("INFO  HTTPS server starting on port %v", srvs.Addr)
			if err := srvs.ListenAndServeTLS(h.TLSCertificate, h.TLSKey); err != http.ErrServerClosed {
				log.Fatalf("ERROR: %v", err)
			}
		}()
	}

	ticker := time.NewTicker(5 * time.Second)
	done := make(chan struct{})

	go func() {
		for {
			select {
			case <-done:
				return

			case <-ticker.C:
				d.sweep()
			}
		}
	}()

	<-shutdown

	ticker.Stop()
	close(done)
}

func (d *dispatcher) sweep() {
	d.auth.Sweep()
}

func (d *dispatcher) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	debug(fmt.Sprintf("%v", r.URL))

	switch strings.ToUpper(r.Method) {
	case http.MethodGet:
		d.get(w, r)
	case http.MethodPost:
		d.post(w, r)
	default:
		http.Error(w, "Invalid request", http.StatusMethodNotAllowed)
	}
}

func (d *dispatcher) authenticate(w http.ResponseWriter, r *http.Request) {
	d.auth.Authenticate(w, r)
}

func (d *dispatcher) authorized(w http.ResponseWriter, r *http.Request, path string) (string, string, bool) {
	return d.auth.Authorized(w, r, path)
}

func (d *dispatcher) user(r *http.Request) string {
	if s, err := d.auth.Session(r); err == nil && s != nil {
		return s.User
	}

	return ""
}

func (d *dispatcher) logout(w http.ResponseWriter, r *http.Request) {
	d.auth.Logout(w, r)
}

func debug(message string) {
	log.Printf("%-5s %s", "DEBUG", message)
}

func info(message string) {
	log.Printf("%-5s %s", "INFO", message)
}

func warn(err error) {
	log.Printf("%-5s %v", "WARN", err)
}
