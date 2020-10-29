module github.com/uhppoted/uhppoted-httpd

go 1.14

require (
	github.com/cristalhq/jwt/v3 v3.0.2
	github.com/google/uuid v1.1.1
	github.com/uhppoted/uhppote-core v0.6.5-0.20200917195138-fc4c9892d764
	github.com/uhppoted/uhppoted-api v0.6.5-0.20200918185449-90cd63ad6cb4
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b // indirect
	golang.org/x/sys v0.0.0-20200812155832-6a926be9bd1d
	golang.org/x/tools v0.0.0-20200821200730-1e23e48ab93b // indirect
)

replace (
	github.com/uhppoted/uhppote-core => ../uhppote-core
	github.com/uhppoted/uhppoted-api => ../uhppoted-api
)

