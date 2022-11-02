package post

import (
	"net/http"

	"github.com/uhppoted/uhppoted-httpd/auth/otp"
	"github.com/uhppoted/uhppoted-httpd/httpd/auth"
	"github.com/uhppoted/uhppoted-httpd/httpd/cookies"
)

func VerifyOTP(w http.ResponseWriter, r *http.Request, auth auth.IAuth) {
	var uid string
	var pwd string
	var OTP string

	keyid := ""
	if cookie, err := r.Cookie(cookies.OTPCookie); err == nil {
		keyid = cookie.Value
	}

	if vars, err := get(r, "uid", "pwd", "otp"); err != nil {
		warnf("OTP", "%v", err)
		http.Error(w, "Error reading request", http.StatusInternalServerError)
		return
	} else {
		uid = vars["uid"]
		pwd = vars["pwd"]
		OTP = vars["otp"]
	}

	if err := auth.Verify(uid, pwd); err != nil {
		warnf("OTP", "%v", err)
		http.Error(w, "Invalid user ID or password", http.StatusBadRequest)
		return
	}

	if err := otp.Validate(uid, keyid, OTP); err != nil {
		warnf("OTP", "%v", err)
		http.Error(w, "Invalid OTP", http.StatusBadRequest)
		return
	}
}
