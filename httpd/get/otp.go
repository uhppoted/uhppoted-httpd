package get

import (
	"compress/gzip"
	"net/http"
	"strings"
	"time"

	"github.com/uhppoted/uhppoted-httpd/auth/otp"
	"github.com/uhppoted/uhppoted-httpd/httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/httpd/cookies"
)

func GenerateOTP(uid string, w http.ResponseWriter, r *http.Request, auth auth.IAuth) {
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

	if err := auth.VerifyAuthHeader(authorization); err != nil {
		warnf("OTP", "%v", err)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// ... generate  OTP
	key := ""
	if cookie, err := r.Cookie(cookies.OTPCookie); err == nil {
		key = cookie.Value
	}

	newkey, qr, err := otp.Get(uid, key)
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
	acceptsGzip := parseHeader(r)
	if acceptsGzip && len(qr) > GZIP_MINIMUM {
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Content-Type", "image/png")

		gz := gzip.NewWriter(w)
		gz.Write(qr)
		gz.Close()
	} else {
		w.Header().Set("Content-Type", "image/png")
		w.Write(qr)
	}
}
