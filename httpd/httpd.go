package httpd

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/fs"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/uhppoted/uhppoted-httpd/httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/httpd/cookies"
	"github.com/uhppoted/uhppoted-httpd/httpd/html"
	"github.com/uhppoted/uhppoted-httpd/log"
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
	withPIN bool
}

func (h *HTTPD) Run(mode types.RunMode, withPIN bool, interrupt chan os.Signal) {
	// ... initialisation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	d := dispatcher{
		fs:      html.HTML,
		auth:    h.AuthProvider,
		context: ctx,
		timeout: h.RequestTimeout,
		mode:    mode,
		withPIN: withPIN,
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

	mux.HandleFunc("/javascript/", d.getJS)

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
			errorf("HTTPD", "Error reading CA certificate file (%v)", err)
			return
		}

		certificates := x509.NewCertPool()
		if !certificates.AppendCertsFromPEM(ca) {
			errorf("HTTPD", "Error parsing CA certificate")
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
			infof("HTTPD", "%v", "terminated")
		}

		cancel()

		if srv != nil {
			if err := srv.Shutdown(context.Background()); err != nil {
				warnf("HTTPD", "HTTP server shutdown error: %w", err)
			}
		}

		if srvs != nil {
			if err := srvs.Shutdown(context.Background()); err != nil {
				warnf("HTTPD", "HTTPS server shutdown error: %w", err)
			}
		}

		close(shutdown)
	}()

	if srv != nil {
		go func() {
			infof("HTTPD", "HTTP  server starting on port %v", srv.Addr)
			if err := srv.ListenAndServe(); err != http.ErrServerClosed {
				fatalf("HTTPD", "%v", err)
			}
		}()
	}

	if srvs != nil {
		go func() {
			infof("HTTPD", "HTTPS server starting on port %v", srvs.Addr)
			if err := srvs.ListenAndServeTLS(h.TLSCertificate, h.TLSKey); err != http.ErrServerClosed {
				fatalf("HTTPD", "%v", err)
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
		debugf("HTTPD", "%-4v %v", r.Method, url)
	} else {
		debugf("HTTPD", "%-4v %v", r.Method, r.URL)
	}

	switch strings.ToUpper(r.Method) {
	case http.MethodHead:
		d.head(w, r)
	case http.MethodGet:
		d.get(w, r)
	case http.MethodPost:
		d.post(w, r)
	case http.MethodDelete:
		d.delete(w, r)
	default:
		http.Error(w, "Invalid request", http.StatusMethodNotAllowed)
	}
}

func (d *dispatcher) authenticated(r *http.Request, w http.ResponseWriter) (string, string, bool) {
	cookie, err := r.Cookie(cookies.SessionCookie)
	if err != nil {
		warnf("HTTPD", "No session cookie in request")
		return "", "", false
	}

	uid, role, cookie2, err := d.auth.Authenticated(cookie)
	if err != nil {
		warnf("HTTPD", "%v", err)
		return "", "", false
	}

	if cookie2 != nil {
		http.SetCookie(w, cookie2)
	}

	return uid, role, true
}

func (d *dispatcher) authorised(uid, role, path string) bool {
	if err := d.auth.Authorised(uid, role, path); err != nil {
		warnf("HTTPD", "%v", err)
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

func debugf(subsystem string, format string, args ...any) {
	if subsystem == "" {
		log.Debugf("%v", args...)
	} else {
		log.Debugf(fmt.Sprintf("%-8v %v", subsystem, format), args...)
	}
}

func infof(subsystem string, format string, args ...any) {
	if subsystem == "" {
		log.Infof("%v", args...)
	} else {
		log.Infof(fmt.Sprintf("%-8v %v", subsystem, format), args...)
	}
}

func warnf(subsystem string, format string, args ...any) {
	if subsystem == "" {
		log.Warnf("%v", args...)
	} else {
		log.Warnf(fmt.Sprintf("%-8v %v", subsystem, format), args...)
	}
}

func errorf(subsystem string, format string, args ...any) {
	if subsystem == "" {
		log.Errorf("%v", args...)
	} else {
		log.Errorf(fmt.Sprintf("%-8v %v", subsystem, format), args...)
	}
}

func fatalf(subsystem string, format string, args ...any) {
	if subsystem == "" {
		log.Fatalf("%v", args...)
	} else {
		log.Fatalf(fmt.Sprintf("%-8v %v", subsystem, format), args...)
	}
}
