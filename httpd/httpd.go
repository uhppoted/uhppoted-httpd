package httpd

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/uhppoted/uhppoted-httpd/httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/httpd/cookies"
	"github.com/uhppoted/uhppoted-httpd/httpd/html"
	"github.com/uhppoted/uhppoted-httpd/types"
)

type HTTPD struct {
	HTML                     string
	HttpEnabled              bool
	HttpsEnabled             bool
	HttpPort                 uint16
	HttpsPort                uint16
	AuthProvider             auth.IAuth
	CACertificate            string
	TLSCertificate           string
	TLSKey                   string
	RequireClientCertificate bool
	RequestTimeout           time.Duration
}

type dispatcher struct {
	fs      fs.FS
	auth    auth.IAuth
	context context.Context
	timeout time.Duration
	mode    types.RunMode
}

func (h *HTTPD) Run(mode types.RunMode, interrupt chan os.Signal) {
	// ... initialisation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	d := dispatcher{
		fs:      html.HTML,
		auth:    h.AuthProvider,
		context: ctx,
		timeout: h.RequestTimeout,
		mode:    mode,
	}

	if h.HTML != "" {
		d.fs = os.DirFS(h.HTML)
	}

	fs := filesystem{
		FileSystem: http.FS(d.fs),
	}

	// ... setup routing
	mux := http.NewServeMux()

	mux.Handle("/css/", http.FileServer(fs))
	mux.Handle("/images/", http.FileServer(fs))
	mux.Handle("/fonts/", http.FileServer(fs))
	mux.Handle("/javascript/", http.FileServer(fs))
	mux.Handle("/manifest.json", http.FileServer(fs))

	mux.HandleFunc("/sys/login.html", d.getNoAuth)
	mux.HandleFunc("/sys/unauthorized.html", d.getNoAuth)
	mux.HandleFunc("/sys/overview.html", d.getWithAuth)
	mux.HandleFunc("/sys/controllers.html", d.getWithAuth)
	mux.HandleFunc("/sys/password.html", d.getWithAuth)
	mux.HandleFunc("/sys/doors.html", d.getWithAuth)
	mux.HandleFunc("/sys/cards.html", d.getWithAuth)
	mux.HandleFunc("/sys/groups.html", d.getWithAuth)
	mux.HandleFunc("/sys/events.html", d.getWithAuth)
	mux.HandleFunc("/sys/logs.html", d.getWithAuth)

	mux.HandleFunc("/authenticate", d.dispatch)
	mux.HandleFunc("/logout", d.dispatch)
	mux.HandleFunc("/password", d.dispatch)
	mux.HandleFunc("/otp", d.dispatch)
	mux.HandleFunc("/interfaces", d.dispatch)
	mux.HandleFunc("/controllers", d.dispatch)
	mux.HandleFunc("/doors", d.dispatch)
	mux.HandleFunc("/cards", d.dispatch)
	mux.HandleFunc("/groups", d.dispatch)
	mux.HandleFunc("/events", d.dispatch)
	mux.HandleFunc("/logs", d.dispatch)
	mux.HandleFunc("/users", d.dispatch)
	mux.HandleFunc("/synchronize/ACL", d.dispatch)
	mux.HandleFunc("/synchronize/datetime", d.dispatch)
	mux.HandleFunc("/synchronize/doors", d.dispatch)

	mux.HandleFunc("/", d.getWithAuth)
	mux.HandleFunc("/usr/", d.getNoAuth)
	mux.HandleFunc("/index.html", d.getNoAuth)

	// ... instantiate servers
	var srv *http.Server
	var srvs *http.Server

	if h.HttpEnabled {
		srv = &http.Server{
			Addr:    fmt.Sprintf(":%v", h.HttpPort),
			Handler: mux,
		}
	}

	if h.HttpsEnabled {
		ca, err := ioutil.ReadFile(h.CACertificate)
		if err != nil {
			log.Printf("%5v Error reading CA certificate file (%v)", "FATAL", err)
			return
		}

		certificates := x509.NewCertPool()
		if !certificates.AppendCertsFromPEM(ca) {
			log.Printf("%5v Error parsing CA certificate", "FATAL")
			return
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
			Addr:      fmt.Sprintf(":%v", h.HttpsPort),
			TLSConfig: &tlsConfig,
			Handler:   mux,
		}
	}

	// ... listen and serve
	shutdown := make(chan struct{})

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)

		select {
		case <-interrupt:
			info("HTTPD", "terminated")
		}

		cancel()

		if srv != nil {
			if err := srv.Shutdown(context.Background()); err != nil {
				warn("HTTPD", fmt.Errorf("HTTP server shutdown error: %w", err))
			}
		}

		if srvs != nil {
			if err := srvs.Shutdown(context.Background()); err != nil {
				warn("HTTPD", fmt.Errorf("HTTPS server shutdown error: %w", err))
			}
		}

		close(shutdown)
	}()

	if srv != nil {
		go func() {
			info("HTTPD", fmt.Sprintf("HTTP  server starting on port %v", srv.Addr))
			if err := srv.ListenAndServe(); err != http.ErrServerClosed {
				log.Panicf("ERROR: %v", err)
			}
		}()
	}

	if srvs != nil {
		go func() {
			info("HTTPD", fmt.Sprintf("HTTPS server starting on port %v", srvs.Addr))
			if err := srvs.ListenAndServeTLS(h.TLSCertificate, h.TLSKey); err != http.ErrServerClosed {
				log.Panicf("ERROR: %v", err)
			}
		}()
	}

	done := make(chan struct{})

	go func() {
		for {
			select {
			case <-done:
				return
			}
		}
	}()

	<-shutdown

	close(done)
}

func (d *dispatcher) dispatch(w http.ResponseWriter, r *http.Request) {
	if url, err := url.QueryUnescape(fmt.Sprintf("%v", r.URL)); err == nil {
		debug(fmt.Sprintf("%-4v %v", r.Method, url))
	} else {
		debug(fmt.Sprintf("%-4v %v", r.Method, r.URL))
	}

	switch strings.ToUpper(r.Method) {
	case http.MethodHead:
		d.head(w, r)
	case http.MethodGet:
		d.get(w, r)
	case http.MethodPost:
		d.post(w, r)
	default:
		http.Error(w, "Invalid request", http.StatusMethodNotAllowed)
	}
}

func (d *dispatcher) authenticated(r *http.Request, w http.ResponseWriter) (string, string, bool) {
	cookie, err := r.Cookie(cookies.SessionCookie)
	if err != nil {
		warn("", fmt.Errorf("No session cookie in request"))
		return "", "", false
	}

	uid, role, cookie2, err := d.auth.Authenticated(cookie)
	if err != nil {
		warn("", err)
		return "", "", false
	}

	if cookie2 != nil {
		http.SetCookie(w, cookie2)
	}

	return uid, role, true
}

func (d *dispatcher) authorised(uid, role, path string) bool {
	if err := d.auth.Authorised(uid, role, path); err != nil {
		warn("", err)
		return false
	}

	return true
}

func (d *dispatcher) unauthenticated(r *http.Request, w http.ResponseWriter) {
	cookies.Clear(w, cookies.SessionCookie, cookies.OTPCookie)
	http.Redirect(w, r, "/sys/login.html", http.StatusFound)
}

func (d *dispatcher) unauthorised(r *http.Request, w http.ResponseWriter) {
	http.Redirect(w, r, "/sys/unauthorized.html", http.StatusFound)
}

func resolve(u *url.URL) (string, error) {
	base, err := url.Parse("/")
	if err != nil {
		return "", err
	}

	return base.ResolveReference(u).EscapedPath(), nil
}

func debug(msg string) {
	log.Printf("%-5s %s", "DEBUG", msg)
}

func debugf(format string, args ...any) {
	f := fmt.Sprintf("%-5v %v", "DEBUG", format)

	log.Printf(f, args...)
}

func info(subsystem string, msg string) {
	if subsystem == "" {
		log.Printf("%-5s %s", "INFO", msg)
	} else {
		log.Printf("%-5s %-8v  %v", "INFO", subsystem, msg)
	}
}

func warn(subsystem string, err error) {
	if subsystem == "" {
		log.Printf("%-5s %v", "WARN", err)
	} else {
		log.Printf("%-5s %-8v  %v", "WARN", subsystem, err)
	}
}
