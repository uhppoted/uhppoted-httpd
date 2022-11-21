package users

import (
	"compress/gzip"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/uhppoted/uhppoted-httpd/auth/otp"
	"github.com/uhppoted/uhppoted-httpd/httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/httpd/cookies"
)

func GenerateOTP(uid, role string, w http.ResponseWriter, r *http.Request, auth auth.IAuth) {
	// ... verify Authorization header
	authorization := ""
	for k, h := range r.Header {
		if strings.TrimSpace(strings.ToLower(k)) == "authorization" {
			for _, v := range h {
				authorization = v
				break
			}
		}
	}

	if err := auth.VerifyAuthHeader(uid, authorization); err != nil {
		warnf("OTP", "%v", err)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// ... generate  OTP
	key := ""
	if cookie, err := r.Cookie(cookies.OTPCookie); err == nil {
		key = cookie.Value
	}

	newkey, expires, qr, err := otp.Get(uid, role, key)
	if err != nil {
		warnf("OTP", "%v", err)
		http.Error(w, "Error generating OTP", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     cookies.OTPCookie,
		Value:    newkey,
		Path:     "/",
		MaxAge:   int((5 * time.Minute).Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		//  Secure:   true,
	})

	// ... reply

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("X-Uhppoted-Httpd-OTP-Expires", fmt.Sprintf("%v", expires.Round(5*time.Second).Seconds()))

	_, acceptsGzip := parseHeader(r)
	if acceptsGzip && len(qr) > GZIP_MINIMUM {
		w.Header().Set("Content-Encoding", "gzip")

		gz := gzip.NewWriter(w)
		gz.Write(qr)
		gz.Close()
	} else {
		w.Write(qr)
	}
}

func VerifyOTP(uid string, role string, w http.ResponseWriter, r *http.Request, auth auth.IAuth) {
	var pwd string
	var OTP string

	keyid := ""
	if cookie, err := r.Cookie(cookies.OTPCookie); err == nil {
		keyid = cookie.Value
	}

	if vars, err := getvars(r, "pwd", "otp"); err != nil {
		warnf("OTP", "%v", err)
		http.Error(w, "Error reading request", http.StatusInternalServerError)
		return
	} else {
		pwd = vars["pwd"]
		OTP = vars["otp"]
	}

	if err := auth.Verify(uid, pwd); err != nil {
		warnf("OTP", "%v", err)
		http.Error(w, "Invalid user ID or password", http.StatusBadRequest)
		return
	}

	if err := otp.Validate(uid, role, keyid, OTP); err != nil {
		warnf("OTP", "%v", err)
		http.Error(w, "Invalid OTP", http.StatusBadRequest)
		return
	}
}

func RevokeOTP(uid string, role string, w http.ResponseWriter, r *http.Request, auth auth.IAuth) {
	// ... verify Authorization header
	authorization := ""
	for k, h := range r.Header {
		if strings.TrimSpace(strings.ToLower(k)) == "authorization" {
			for _, v := range h {
				authorization = v
				break
			}
		}
	}

	if err := auth.VerifyAuthHeader(uid, authorization); err != nil {
		warnf("OTP", "%v", err)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// ... revoke OTP
	if err := otp.Revoke(uid, role); err != nil {
		warnf("OTP", "%v", err)
		http.Error(w, "Invalid OTP", http.StatusBadRequest)
		return
	}
}
