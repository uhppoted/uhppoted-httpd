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

	"github.com/google/uuid"

	"github.com/uhppoted/uhppoted-httpd/auth"
)

const (
	LoginCookie   = "uhppoted-httpd-login"
	SessionCookie = "uhppoted-httpd-session"
)

type HTTPD struct {
	Dir                      string
	AuthProvider             auth.IAuth
	CookieMaxAge             int
	HTTPEnabled              bool
	HTTPSEnabled             bool
	CACertificate            string
	TLSCertificate           string
	TLSKey                   string
	RequireClientCertificate bool
}

type session struct {
	id   uuid.UUID
	user string
}

type dispatcher struct {
	root         string
	fs           http.Handler
	auth         auth.IAuth
	cookieMaxAge int
	logins       map[uuid.UUID]bool
	sessions     map[uuid.UUID]*session
}

func (h *HTTPD) Run() {
	fs := httpdFileSystem{
		FileSystem: http.Dir(h.Dir),
	}

	d := dispatcher{
		root:         h.Dir,
		fs:           http.FileServer(fs),
		auth:         h.AuthProvider,
		cookieMaxAge: h.CookieMaxAge,
		logins:       map[uuid.UUID]bool{},
		sessions:     map[uuid.UUID]*session{},
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

	<-shutdown
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

func (d *dispatcher) authorised(r *http.Request, path string) bool {
	if path == "/login.html" {
		return true
	}

	if path == "/unauthorized.html" {
		return true
	}

	if strings.HasSuffix(path, ".html") {
		cookie, err := r.Cookie(SessionCookie)
		if err != nil {
			warn(fmt.Errorf("No JWT cookie in request"))
			return false
		}

		if err := d.auth.Authorized(cookie.Value, path); err != nil {
			warn(err)
			return false
		}

		session, err := d.session(r)
		if err != nil {
			warn(err)
			return false
		}

		if session == nil {
			warn(fmt.Errorf("No extant session for request"))
			return false
		}
	}

	return true
}

func (d *dispatcher) session(r *http.Request) (*session, error) {
	cookie, err := r.Cookie(SessionCookie)
	if err != nil {
		return nil, err
	}

	sid, err := d.auth.GetSessionId(cookie.Value)
	if err != nil {
		return nil, err
	}

	if sid == nil {
		return nil, fmt.Errorf("Invalid session ID (%v)", sid)
	}

	s, ok := d.sessions[*sid]
	if !ok {
		return nil, fmt.Errorf("No extant session for session ID '%v'", *sid)
	}

	return s, nil
}

func (d *dispatcher) unauthorized(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/unauthorized.html", http.StatusFound)
}

func (d *dispatcher) logout(w http.ResponseWriter, r *http.Request) {
	if s, _ := d.session(r); s != nil {
		delete(d.sessions, s.id)
	}

	http.Redirect(w, r, "/index.html", http.StatusFound)
}

func authorize(header []string) error {
	if len(header) == 0 {
		return fmt.Errorf("Empty 'Authorization' header")
	}

	return nil
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
