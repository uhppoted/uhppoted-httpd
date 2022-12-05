# CHANGELOG

## [Unreleased]

1. Added support for OTP as a password alternative on login. Please see security 
   note in [README](https://github.com/uhppoted/uhppoted-httpd#notes).
2. Added optonal [OKSolar](https://meat.io/oksolar) palette.
3. Implemented `config` command to display system configuration.

### Changed
1. Reworked 'change password' to use Authorization header
2. Locked user login after too many failed attempts
3. Removed legacy support for users in auth.json
4. Updated _systemd_ unit file to wait for `network-online.target`

## [0.8.2](https://github.com/uhppoted/uhppoted-httpd/releases/tag/v0.8.2) - 2022-10-14

### Changed
1. Bumped Go version to 1.19
2. Added _eslint_ setup to README


## [0.8.1](https://github.com/uhppoted/uhppoted-httpd/releases/tag/v0.8.1) - 2022-08-01

### Changed
1. Stored 'missing events' to avoid stalling event retrieval.
2. Fixed missing 'onMore' handler.
3. Overrode Chrome's autofill setting for login UID field.
4. Displayed the created admin user ID and password at the end of the 'daemonize' output.
5. Added 'fonts' folder to embedded HTML file system.


## [0.8.0](https://github.com/uhppoted/uhppoted-httpd/releases/tag/v0.8.0) - 2022-07-01

1. Initial release

