package httpd

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/uhppoted/uhppoted-httpd/httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/httpd/html"
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
}

type dispatcher struct {
	root    string
	auth    auth.IAuth
	context context.Context
	timeout time.Duration
}

const (
	SettingsCookie = "uhppoted-settings"
)

func (h *HTTPD) Run() {
	// ... initialisation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	d := dispatcher{
		root:    "", // h.Dir,
		auth:    h.AuthProvider,
		context: ctx,
		timeout: h.RequestTimeout,
	}

	//	fs := filesystem{
	//		http.FS(os.DirFS(h.Dir)),
	//	}

	fs := filesystem{
		FileSystem: http.FS(html.STATIC),
	}

	// ... setup routing
	mux := http.NewServeMux()

	mux.Handle("/css/", http.FileServer(fs))
	mux.Handle("/images/", http.FileServer(fs))
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
	mux.HandleFunc("/interfaces", d.dispatch)
	mux.HandleFunc("/controllers", d.dispatch)
	mux.HandleFunc("/doors", d.dispatch)
	mux.HandleFunc("/cards", d.dispatch)
	mux.HandleFunc("/groups", d.dispatch)
	mux.HandleFunc("/events", d.dispatch)
	mux.HandleFunc("/logs", d.dispatch)
	mux.HandleFunc("/users", d.dispatch)

	mux.HandleFunc("/", d.getWithAuth)
	mux.HandleFunc("/usr/", d.getNoAuth)
	mux.HandleFunc("/index.html", d.getNoAuth)

	// ... instantiate servers
	var srv *http.Server
	var srvs *http.Server

	if h.HTTPEnabled {
		srv = &http.Server{
			Addr:    ":8080",
			Handler: mux,
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
			Handler:   mux,
		}
	}

	// ... listen and serve
	shutdown := make(chan struct{})

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)

		<-sigint

		cancel()

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
	cookie, err := r.Cookie(auth.SessionCookie)
	if err != nil {
		warn(fmt.Errorf("No session cookie in request"))
		return "", "", false
	}

	uid, role, cookie2, err := d.auth.Authenticated(cookie)
	if err != nil {
		warn(err)
		return "", "", false
	}

	if cookie2 != nil {
		http.SetCookie(w, cookie2)
	}

	return uid, role, true
}

func (d *dispatcher) authorised(uid, role, path string) bool {
	if err := d.auth.Authorised(uid, role, path); err != nil {
		warn(err)
		return false
	}

	return true
}

func (d *dispatcher) unauthenticated(r *http.Request, w http.ResponseWriter) {
	clear(auth.SessionCookie, w)

	http.Redirect(w, r, "/sys/login.html", http.StatusFound)
}

func (d *dispatcher) unauthorised(r *http.Request, w http.ResponseWriter) {
	http.Redirect(w, r, "/sys/unauthorized.html", http.StatusFound)
}

// cf. https://stackoverflow.com/questions/27671061/how-to-delete-cookie
func clear(cookie string, w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookie,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		//	Secure:   true,
	})
}

func resolve(u *url.URL) (string, error) {
	base, err := url.Parse("/")
	if err != nil {
		return "", err
	}

	return base.ResolveReference(u).EscapedPath(), nil
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
