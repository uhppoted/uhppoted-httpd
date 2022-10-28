package post

import (
	"net/http"

	"github.com/uhppoted/uhppoted-httpd/auth/otp"
	"github.com/uhppoted/uhppoted-httpd/httpd/auth"
)

func VerifyOTP(w http.ResponseWriter, r *http.Request, auth auth.IAuth) {
	var uid string
	var pwd string
	var otp1 string
	var otp2 string

	_, acceptsGzip := parseHeader(r)

	if vars, err := get(r, "uid", "pwd", "otp1", "otp2"); err != nil {
		warnf("OTP", "%v", err)
		http.Error(w, "Error reading request", http.StatusInternalServerError)
		return
	} else {
		uid = vars["uid"]
		pwd = vars["pwd"]
		otp1 = vars["otp1"]
		otp2 = vars["otp2"]
	}

	if err := auth.Verify(uid, pwd); err != nil {
		warnf("OTP", "%v", err)
		http.Error(w, "Invalid user ID or password", http.StatusBadRequest)
		return
	}

	if err := otp.Validate(uid, otp1); err != nil {
		warnf("OTP", "%v", err)
		http.Error(w, "Invalid OTP", http.StatusBadRequest)
		return
	}

	if err := otp.Validate(uid, otp2); err != nil {
		warnf("OTP", "%v", err)
		http.Error(w, "Invalid OTP", http.StatusBadRequest)
		return
	}

	if acceptsGzip {

	}

	http.Error(w, "(work in progress)", http.StatusInternalServerError)
}
