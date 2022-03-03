![build](https://github.com/uhppoted/uhppoted-httpd/workflows/build/badge.svg)

# uhppoted-httpd

`uhppoted-httpd` implements an HTTP server that provides a browser based user interface for an access control system based
on UHPPOTE TCP/IP controllers. It is intended to supplement the existing command line tools and application integrations.

## Status

_In development_

Supported operating systems:
- Linux
- MacOS
- Windows
- ARM7 _(e.g. RaspberryPi)_

## Raison d'être

** CAVEAT EMPTOR **

`uhppoted-httpd`, at this stage in its career, is primarily intended for use in validating the design and
implementation of the other `uhppoted` components when integrated into a working system. It is also intended
for exploring some alternative ideas around user interfaces and system architectures.

Having said which, it does provide a functional and usable user interface for managing a small'ish access
control system. Be warned though - the out-of-the-box look and feel is (deliberately) workaday, low key and
plain and is very definitely intended as a basis for your own customisation with your own logos, themes and
functionality.

## Releases

| *Version* | *Description*                                                                             |
| --------- | ----------------------------------------------------------------------------------------- |
|           |                                                                                           |
|           |                                                                                           |

## Installation

Executables for all the supported operating systems are packaged in the [releases](https://github.com/uhppoted/uhppoted-httpd/releases):

The release tarballs contain the executables for all the operating systems - OS specific tarballs with all the _uhppoted_ components can be found in [uhpppoted](https://github.com/uhppoted/uhppoted/releases) releases.

Installation is straightforward - download the archive and extract it to a directory of your choice. To install `uhppoted-httpd` as a system service:
```
   cd <uhppote directory>
   sudo uhppoted-httpd daemonize
```

`uhppoted-httpd help` will list the available commands and associated options (documented below).

The `daemonize` command will create all the necessary files for `uhppoted-httpd` if they do not exist already:

- `uhppoted.conf`
- access lists
- GRULES files
- HTML files

### Building from source

Assuming you have `Go`, `make` and `eslint` installed:

```
git clone https://github.com/uhppoted/uhppoted-httpd.git
cd uhppoted-httpd
make build
```

If you prefer not to use `make`:
```
git clone https://github.com/uhppoted/uhppoted-httpd.git
cd uhppoted-httpd
mkdir bin
go build -trimpath -o bin ./...
```

The above commands build the `'uhppoted-httpd` executable to the `bin` directory.


#### Dependencies

| *Dependency*                                                            | *Description*                        |
| ----------------------------------------------------------------------- | -------------------------------------|
| [uhppote-core](https://github.com/uhppoted/uhppote-core)                | Device level API implementation      |
| [uhppoted-lib](https://github.com/uhppoted/uhppoted-lib)                | common API for external applications |
| [jwt/v3](https://github.com/cristalhq/jwt/v3)                           | JWT implementation                   |
| [grule-rule-engine](https://github.com/hyperjumptech/grule-rule-engine) | Rules engine                         |
| github.com/google/uuid                                                  | UUID type implementation             |
| golang.org/x/sys                                                        | (for Windows service integration)    |

## uhppoted-httpd

Usage: ```uhppoted-httpd <command> <options>```

Supported commands:

- `help`
- `version`
- `config`
- `run`
- `console`
- `daemonize`
- `undaemonize`

Defaults to `run` if the command it not provided i.e. ```uhppoted-httpd <options>``` is equivalent to 
```uhppoted-httpd run <options>```.

### `run`

Runs the `uhppoted-httpd` HTTP server. Default command, intended for use as a system service that runs in the 
background. 

Command line:

` uhppoted-httpd [--debug] [--console] [--config <file>] `

```
  --config      Sets the uhppoted.conf file to use for controller configurations. 
                Defaults to the communal uhppoted.conf file shared by all the uhppoted modules.
  --console     Runs the HTTP server endpoint as a console application, logging events to the console.
  --debug       Displays verbose debugging information, in particular the communications with the 
                UHPPOTE controllers
```

### `daemonize`

Registers `uhppoted-httpd` as a system service that will be started on system boot. The command creates the necessary
system specific service configuration files and service manager entries. On Linux it defaults to using the 
`uhppoted:uhppoted` user:group - this can be changed with the `--user` option

Command line:

`uhppoted-httpd daemonize [--user <user>]`

### `undaemonize`

Unregisters `uhppoted-httpd` as a system service, but does not delete any created log or configuration files. 

Command line:

`uhppoted-httpd undaemonize `

## Supporting files

### `uhppoted.conf`

`uhppoted.conf` is the communal configuration file shared by all the `uhppoted` project modules and is (or will 
eventually be) documented in [uhppoted](https://github.com/uhppoted/uhppoted). `uhppoted-httpd` requires:
- the _HTTPD_ section to define the configuration for the HTTP server
- the _devices_ section to resolve non-local controller IP addresses and door to controller door identities.

The `daemonize` command will create a `uhppoted.conf` file if one does not exist, or update the existing file
with the default configuration.

### HTML files

### `auth.json`

### `ACL.grl`

### GRULES files

### JSON files



