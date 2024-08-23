# CHANGELOG

## Unreleased

### Added
1. Implemented TCP/IP transport support.
2. On startup, redirects to 'setup' page to create an admin user if none exists.

### Updated
1. Removed creation of admin user in _daemonize_.
2. Added ADMIN role check to default _grules_ rulesets.
3. _admin_ role is configurable.
4. Fixed bug in authentication that allowed deleted user to login unless system
   had been restarted.
5. Updated to Go 1.23.


## [0.8.8](https://github.com/uhppoted/uhppoted-httpd/releases/tag/v0.8.8) - 2024-03-27

### Added
1. Added public Docker image to ghcr.io.

### Updated
1. Bumped Go version to 1.22.


## [0.8.7](https://github.com/uhppoted/uhppoted-httpd/releases/tag/v0.8.7) - 2023-12-01

### Added
1. Added _passcodes_ field to _doors_ page to set override passcodes.

### Updated
1. Added NodeJS v18.19.1 and v20.11.1 to the CI build matrix


## [0.8.6](https://github.com/uhppoted/uhppoted-httpd/releases/tag/v0.8.6) - 2023-08-30

### Added
1. Implemented door keypad activation/deactivation to _doors_ page and updatable doors record.

### Updated
1. Renamed _master_ branch to _main_ in line with current development practice.
2. Replaced os.Rename with robust implementation for moving files between file systems.


## [0.8.5](https://github.com/uhppoted/uhppoted-httpd/releases/tag/v0.8.5) - 2023-06-13

### Added
1. Added controller interlock mode to system page and updatable controller record.


## [0.8.4](https://github.com/uhppoted/uhppoted-httpd/releases/tag/v0.8.4) - 2023-03-17

### Added
1. `doc.go` package overview documentation.
2. Added PIN support for card keypad PIN

### Updated
1. Fixed initial round of _staticcheck_ lint errors and added _staticcheck_ to
   CI build.


## [0.8.3](https://github.com/uhppoted/uhppoted-httpd/releases/tag/v0.8.3) - 2022-11-16

### Added
1. Added support for OTP as a password alternative on login. Please see security 
   note in [README](https://github.com/uhppoted/uhppoted-httpd#notes).
2. Added optonal [OKSolar](https://meat.io/oksolar) palette.
3. Implemented `config` command to display system configuration.
4. Added ARM64 to release build artifacts

### Changed
1. Reworked 'change password' to use Authorization header
2. Locked user login after too many failed attempts
3. Removed legacy support for users in auth.json
4. Updated _systemd_ unit file to wait for `network-online.target`
5. Reworked lockfile to use `flock` _syscall_.
6. Moved default lockfile to system _temp_ folder.


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

