module github.com/uhppoted/uhppoted-httpd

go 1.14

require (
	github.com/cristalhq/jwt/v3 v3.0.2
	github.com/google/uuid v1.1.1
	github.com/uhppoted/uhppote-core v0.6.4
	github.com/uhppoted/uhppoted-api v0.6.4
	golang.org/x/sys v0.0.0-20200812155832-6a926be9bd1d
)

replace github.com/uhppoted/uhppoted-api => ../uhppoted-api
