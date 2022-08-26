# fvpn

fvpn - is a Forest VPN CLI client for macOS, Linux, and Windows.

# How-to

```
NAME:
   fvpn - fast, secure, and modern VPN

USAGE:
   fvpn [global options] command [command options] [arguments...]

COMMANDS:
   account   Manage your account
   state     Control the state of connection
   location  Manage locations
   version   Show the version of ForestVPN CLI
   help, h   Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help (default: false)
```

## Usage example
Register an account:
```
fvpn account register
```
Or login into existing one:
```
fvpn account login
```
See available locations:
```
fvpn location ls
```
Choose or change the location: 
```
fvpn location set ${CITY}
```
Connect to the chosen location:
```
fvpn state up
```
Disconnect from the chosen location:
```
fvpn state down
```

# Installation

## macOS
```
brew install forestvpn/core/fvpn
```
Make sure you have [Homebrew installed](https://docs.brew.sh/Installation) on your Mac.

## Debian/Ubuntu

```
wget -q https://github.com/forestvpn/cli/releases/latest/download/fvpn_linux_amd64.deb && dpkg -i fvpn_linux_amd64.deb
```

## Fedora

```
wget -q https://github.com/forestvpn/cli/releases/latest/download/fvpn_linux_amd64.rpm && dnf install fvpn_linux_amd64.rpm
```

## Alpine

```
wget -q https://github.com/forestvpn/cli/releases/latest/download/fvpn_linux_amd64.apk && apk add fvpn_linux_amd64.apk --allow-untrusted
```

## Linux

```
wget -q https://github.com/forestvpn/cli/releases/latest/download/fvpn_linux_amd64.tar.gz && tar -xf fvpn_linux_amd64.tar.gz -C /usr/local/bin/
```

# Dependencies

- net-tools
- wireguard-tools

# Docs

fvpn consists of various pacakges:

- [actions](https://github.com/forestvpn/cli/tree/main/src/actions#readme) is a high-level package that implements Actions according to https://cli.urfave.org/v2
- [api](https://github.com/forestvpn/cli/tree/main/src/api#readme) is a package that uses [api-client-go](https://github.com/forestvpn/api-client-go) to query [wgrest API](https://github.com/suquant/wgrest)
- [auth](https://github.com/forestvpn/cli/tree/main/src/auth#readme) is a package containing authentication logic built around [Firebase REST API](https://firebase.google.com/docs/reference/rest)
- [cmd](https://github.com/forestvpn/cli/tree/main/src/cmd#readme) is fvpn's entry point followed by https://cli.urfave.org/v2 pattern
- [utils](https://github.com/forestvpn/cli/tree/main/src/utils#readme) is a package that provides helper functions to  work with local filesystem, networking, etc

# Credits:

- ForestVPN.com [Free VPN](https://forestvpn.com) for all
- SpaceV.net [VPN for teams](https://spacev.net)
