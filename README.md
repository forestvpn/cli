# fvpn

fvpn - is a Forest VPN CLI client for Linux distributions.

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

# Installation

## macOS

```
brew install forestvpn/cli/fvpn
```

## Debian/Ubuntu

```
curl -L https://github.com/forestvpn/cli/releases/latest/download/fvpn_linux_amd64.deb -o fvpn_linux_amd64.deb
dpkg -i fvpn_linux_amd64.deb
```

## Fedora

```
curl -L https://github.com/forestvpn/cli/releases/latest/download/fvpn_linux_amd64.rpm -o fvpn_linux_amd64.rpm
dnf install fvpn_linux_amd64.rpm
```

## Alpine

```
curl -L https://github.com/forestvpn/cli/releases/latest/download/fvpn_linux_amd64.apk -o fvpn_linux_amd64.apk
apk add fvpn_linux_amd64.apk --allow-untrusted
```

## Linux

```
curl -L https://github.com/forestvpn/cli/releases/latest/download/fvpn_linux_amd64.tar.gz -o fvpn_linux_amd64.tar.gz
tar -xf fvpn_linux_amd64.tar.gz -C /usr/local/bin/
```