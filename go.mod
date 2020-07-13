module github.com/uhppoted/uhppoted-httpd

go 1.14

require (
	github.com/cristalhq/jwt/v3 v3.0.0
	github.com/google/uuid v1.1.1
	github.com/uhppoted/uhppote-core v0.6.3
	github.com/uhppoted/uhppoted-api v0.6.3
	golang.org/x/sys v0.0.0-20200223170610-d5e6a3e2c0ae
)

replace github.com/uhppoted/uhppoted-api => ../uhppoted-api
