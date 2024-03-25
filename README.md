![build](https://github.com/uhppoted/uhppoted-httpd/workflows/build/badge.svg)
![build](https://github.com/uhppoted/uhppoted-httpd/workflows/ghcr/badge.svg)

# uhppoted-httpd

`uhppoted-httpd` implements an HTTP server that provides a browser based user interface for an access control system based
on UHPPOTE TCP/IP controllers. It is intended to supplement the existing command line tools and application integrations.

## Status

Supported operating systems:
- Linux
- MacOS
- Windows
- ARM7 _(e.g. RaspberryPi)_

## Raison d'être

_CAVEAT EMPTOR_

Although `uhppoted-httpd` does provide a functional and usable user interface for managing a small'ish access
control system, the out-of-the-box look and feel is (deliberately) workaday, low key and plain with the intention
of being a base for your own customisation (with your own logos, themes, functionality, etc) rather than a 
finished, shippable product.

Also, please be aware that at this stage in its career, it is primarily a testbed for validating the design and
implementation of the other `uhppoted` components when integrated into a working system. It is also intended to
become a platform for exploring some alternative ideas around user interfaces and system architectures.

## Release Notes

### Current Release

**[v0.8.8](https://github.com/uhppoted/uhppoted-httpd/releases/tag/v0.8.8) - 2024-03-28**

1. Added public Docker image to ghcr.io.
2. Bumped Go version to 1.22.


## Installation

Executables for all the supported operating systems are packaged in the [releases](https://github.com/uhppoted/uhppoted-httpd/releases):

The release tarballs contain the executables for all the operating systems - OS specific tarballs with all the _uhppoted_ components can be found in [uhpppoted](https://github.com/uhppoted/uhppoted/releases) releases.

Installation is straightforward - download the archive and extract it to a directory of your choice. To install `uhppoted-httpd` as a system service:
```
   cd <uhppoted directory>
   sudo uhppoted-httpd daemonize
```

`uhppoted-httpd help` will list the available commands and associated options (documented below).

The `daemonize` command will create all the necessary files for `uhppoted-httpd` if they do not exist already:

- `uhppoted.conf`
- access lists
- GRULES files
- HTML files


### Docker

A public _Docker_ image is published to [ghcr.io](https://github.com/uhppoted?tab=packages&repo_name=uhppoted-httpd). 

The image is configured to use the `/usr/local/etc/uhppoted/uhppoted.conf` file for configuration information.

#### `docker compose`

A sample Docker `compose` configuration is provided in the [`scripts/docker/compose`](scripts/docker/compose) folder. 

To run the example, download and extract the [zipped](script/docker/compose.zip) scripts and supporting files into folder
of your choice and then:
```
cd <compose folder>
docker compose up
```

And open URL http://localhost:8080 in your browser of choice.

The default image is configured for HTTP only. The supplied compose.yml file uses _bind_ mounts to the local folder to
override the default configuration and HTML.

Alternatively, copy the uhppoted.conf file, TLS keys and certificates and HTML to a Docker volume and remove the bind mounts
from _compose.yml_. The expected folder structure is:
```
/
  usr
    local
      etc
        uhppoted
          - uhppoted.conf
          rest
            - ca.cert
            - uhppoted.key
            - uhppoted.cert
            grules
              - ...
            system
              - ...
            html
              - ...
```

#### `docker run`

To start a REST server using Docker `run`:
```
docker pull ghcr.io/uhppoted/httpd:latest
docker run --publish 8080:8080 --publish 8443:8443 --name httpd --mount source=uhppoted,target=/var/uhppoted --rm ghcr.io/uhppoted/httpd
```

And open URL http://localhost:8080 in your browser of choice.


#### `docker build`

For inclusion in a Dockerfile:
```
FROM ghcr.io/uhppoted/httpd:latest
```


### Building from source

Required tools:
- [Go 1.21+](https://go.dev)
- [sass](https://sass-lang.com)
- make (optional but recommended)
- [eslint](https://eslint.org) (optional but recommended)
- [eslint-config-standard](https://www.npmjs.com/package/eslint-config-standard) (optional but recommended)

**NOTES:**

1. `apt install sass` on Ubuntu installs `ruby-sass` which was marked **[obsolete](https://sass-lang.com/ruby-sass)**
in 2019. Please follow the installation instructions on the [Sass homepage](https://sass-lang.com) to install
the current version._

2. The _make_ build uses `eslint` and `eslint_config_standard`. `eslint_config_standard` is a **dev** dependency and
should be installed locally in the project:

    * Initial project setup:
```
git clone https://github.com/uhppoted/uhppoted-httpd.git
cd uhppoted-httpd
npm install eslint-config-standard
```
   * To build using the included Makefile:
```
cd uhppoted-httpd
make build
```
   * Without using `make`:
```
cd uhppoted-httpd
sass --no-source-map sass/themes/light:httpd/html/css/default
sass --no-source-map sass/themes/light:httpd/html/css/light
sass --no-source-map sass/themes/dark:httpd/html/css/dark
cp httpd/html/images/light/* httpd/html/images/default
go build -trimpath -o bin/ ./...
```

The above commands build the `uhppoted-httpd` executable to the `bin` directory.


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
- `run`
- `daemonize`
- `undaemonize`
- `config`

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
  --lockfile    (optional) Lockfile used to prevent running multiple copies of the _uhppoted-httpd_ service. 
                Defaults to _uhppoted-httpd.pid" (in the system _temp_ folder) if not provided.
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

`uhppoted-httpd undaemonize`

### `config`

Displays the current system configuration. Primarily intended as a convenience for scripts but can also be used to
create a _uhppoted.conf_ file by directing the output to a file (e.g. `uhppoted-http config > /etc/uhppoted/uhppoted.conf`)

Command line:

`uhppoted-httpd config`


## Supporting files

### `uhppoted.conf`

`uhppoted.conf` is the communal configuration file shared by all the `uhppoted` project modules and is (or will 
eventually be) documented in [uhppoted](https://github.com/uhppoted/uhppoted). The `daemonize` command will 
create a `uhppoted.conf` file if one does not exist, or update the existing file with the default configuration.

The configuration for `uhppoted-httpd` is defined in the [_HTTPD_](https://github.com/uhppoted/uhppoted-httpd/blob/master/documentation/uhppoted.conf.md) section.

### HTML files

By default, the static files for the user interface are served from a file system embedded in the application
executable. For customisation, the static files can be relocated to an external folder, as described here:

- [HTML](https://github.com/uhppoted/uhppoted-httpd/blob/master/documentation/HTML.md)

### `auth.json`

Coarse-grained authorisation for HTTP request is set by the entries in the `auth.json` file, which maps URLs and
user roles to GET/POST rights. Detailed description of the file can be found here:

- [auth.json](https://github.com/uhppoted/uhppoted-httpd/blob/master/documentation/auth.json.md)


### `acl.grl`

The `acl.grl` file implements rule based access for cards to supplement the relatively simple grid-based access
control supported by the combination of card + groups + doors. The `acl.grl` file is documented in more detail 
[here](https://github.com/uhppoted/uhppoted-httpd/blob/master/documentation/acl.grl.md).

### GRULES files

The _grules_ files implement rule based fine-grained authorisation for view, create, update and delete operations
on individual entities.. The _grules_ files are documented in more detail [here](https://github.com/uhppoted/uhppoted-httpd/blob/master/documentation/grules.md).

### JSON files

The system data is (currently) stored as a set of JSON files, described (https://github.com/uhppoted/uhppoted-httpd/blob/master/documentation/db.md).

## Notes

1. `uhppoted-http` supports using OTP as an **alternative** to password based login. On that grounds that the 
   most asked question so far has been _"I've forgotten the admin password, how do I ..."_ it seems that once the system
   is setup and configured most users access it sufficiently infrequently for a secure password to be onerous. Login with
   OTP is a convenient alternative using something like e.g. Google Authenticator. Please note that is is less secure than
   using password-only access (of necessity, OTP secret keys are stored in plaintext on the server) so OTP should only be
   enabled if the server is secured.





