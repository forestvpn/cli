#!/bin/sh

main (){
    URL="https://github.com/forestvpn/cli/releases/latest"
    OS=""
	VERSION=""
	PACKAGETYPE=""
	APT_SYSTEMCTL_START=false # Only needs to be true for Kali
	TRACK="${TRACK:-stable}"
    ARCH=$(uname -m)

    if [ -f /etc/os-release ]; then
		# /etc/os-release populates a number of shell variables. We care about the following:
		#  - ID: the short name of the OS (e.g. "debian", "freebsd")
		#  - VERSION_ID: the numeric release version for the OS, if any (e.g. "18.04")
		#  - VERSION_CODENAME: the codename of the OS release, if any (e.g. "buster")
		#  - UBUNTU_CODENAME: if it exists, use instead of VERSION_CODENAME
		. /etc/os-release
		case "$ID" in
			ubuntu|pop|neon|zorin)
				OS="ubuntu"
				PACKAGETYPE="apt"
				;;
			debian)
				OS="$ID"
				VERSION="$VERSION_CODENAME"
				PACKAGETYPE="apt"
				;;
			linuxmint)
				if [ "${UBUNTU_CODENAME:-}" != "" ]; then
				    OS="ubuntu"
				    VERSION="$UBUNTU_CODENAME"
				elif [ "${DEBIAN_CODENAME:-}" != "" ]; then
				    OS="debian"
				    VERSION="$DEBIAN_CODENAME"
				else
				    OS="ubuntu"
				    VERSION="$VERSION_CODENAME"
				fi
				PACKAGETYPE="apt"
				;;
			elementary)
				OS="ubuntu"
				VERSION="$UBUNTU_CODENAME"
				PACKAGETYPE="apt"
				;;
			parrot)
				OS="debian"
				PACKAGETYPE="apt"
				if [ "$VERSION_ID" -lt 5 ]; then
					VERSION="buster"
				else
					VERSION="bullseye"
				fi
				;;
			raspbian)
				OS="$ID"
				VERSION="$VERSION_CODENAME"
				PACKAGETYPE="apt"
				;;
			kali)
				OS="debian"
				PACKAGETYPE="apt"
				YEAR="$(echo "$VERSION_ID" | cut -f1 -d.)"
				APT_SYSTEMCTL_START=true
				# Third-party keyrings became the preferred method of
				# installation in Debian 11 (Bullseye), which Kali switched
				# to in roughly 2021.x releases
				if [ "$YEAR" -lt 2021 ]; then
					# Kali VERSION_ID is "kali-rolling", which isn't distinguishing
					VERSION="buster"
				else
					VERSION="bullseye"
				fi
				;;
			centos)
				OS="$ID"
				VERSION="$VERSION_ID"
				PACKAGETYPE="dnf"
				if [ "$VERSION" = "7" ]; then
					PACKAGETYPE="yum"
				fi
				;;
			ol)
				OS="oracle"
				VERSION="$(echo "$VERSION_ID" | cut -f1 -d.)"
				PACKAGETYPE="dnf"
				if [ "$VERSION" = "7" ]; then
					PACKAGETYPE="yum"
				fi
				;;
			rhel)
				OS="$ID"
				VERSION="$(echo "$VERSION_ID" | cut -f1 -d.)"
				PACKAGETYPE="dnf"
				if [ "$VERSION" = "7" ]; then
					PACKAGETYPE="yum"
				fi
				;;
			fedora)
				OS="$ID"
				VERSION=""
				PACKAGETYPE="dnf"
				;;
			rocky|almalinux|nobara)
				OS="fedora"
				VERSION=""
				PACKAGETYPE="dnf"
				;;
			amzn)
				OS="amazon-linux"
				VERSION="$VERSION_ID"
				PACKAGETYPE="yum"
				;;
			xenenterprise)
				OS="centos"
				VERSION="$(echo "$VERSION_ID" | cut -f1 -d.)"
				PACKAGETYPE="yum"
				;;
			alpine)
				OS="$ID"
				VERSION="$VERSION_ID"
				PACKAGETYPE="apk"
				;;
			osmc)
				OS="debian"
				PACKAGETYPE="apt"
				VERSION="bullseye"
				;;
		esac
	fi

    # If we failed to detect something through os-release, consult
	# uname and try to infer things from that.
	if [ -z "$OS" ]; then
		if type uname >/dev/null 2>&1; then
			case "$(uname)" in
				Darwin)
					OS="macos"
					VERSION="$(sw_vers -productVersion | cut -f1-2 -d.)"
					PACKAGETYPE="brew"
					;;
				Linux)
					OS="other-linux"
					VERSION=""
					PACKAGETYPE=""
					;;
			esac
		fi
	fi

    # Step 2: having detected an OS we support, is it one of the
	# versions we support?
	OS_UNSUPPORTED=
	case "$OS" in
		opensuse)
		    OS_UNSUPPORTED=1
			;;
		fedora)
			# All versions supported, no version checking required.
			;;
		arch)
			OS_UNSUPPORTED=1
			;;
		manjaro)
			OS_UNSUPPORTED=1
			;;
		void)
			OS_UNSUPPORTED=1
			;;
		gentoo)
			OS_UNSUPPORTED=1
			;;
		freebsd)
			OS_UNSUPPORTED=1
			;;
        nixos)
            OS_UNSUPPORTED=1
            ;;
		openbsd)
			OS_UNSUPPORTED=1
			;;
		macos)
			# All versions supported, no version checking required.
			;;
		other-linux)
			OS_UNSUPPORTED=1
			;;
		*)
			OS_UNSUPPORTED=1
			;;
	esac
	if [ "$OS_UNSUPPORTED" = "1" ]; then
		case "$OS" in
			other-linux)
				echo "Couldn't determine what kind of Linux is running."
				echo "You could try the static binaries at:"
				echo "https://github.com/forestvpn/cli/releases/latest"
				;;
			"")
				echo "Couldn't determine what operating system you're running."
				;;
			*)
				echo "$OS $VERSION isn't supported by this script yet."
				;;
		esac
		if type uname >/dev/null 2>&1; then
			echo "UNAME=$(uname -a)"
		else
			echo "UNAME="
		fi
		echo
		if [ -f /etc/os-release ]; then
			cat /etc/os-release
		else
			echo "No /etc/os-release"
		fi
		exit 1
	fi

    # Step 3: work out if we can run privileged commands, and if so,
	# how.
	CAN_ROOT=
	SUDO=
	if [ "$(id -u)" = 0 ]; then
		CAN_ROOT=1
		SUDO=""
	elif type sudo >/dev/null; then
		CAN_ROOT=1
		SUDO="sudo"
	elif type doas >/dev/null; then
		CAN_ROOT=1
		SUDO="doas"
	fi
	if [ "$CAN_ROOT" != "1" ]; then
		echo "This installer needs to run commands as root."
		echo "We tried looking for 'sudo' and 'doas', but couldn't find them."
		echo "Either re-run this script as root, or set up sudo/doas."
		exit 1
	fi

    # Step 4: run the installation.
	echo "Installing ForestVPN CLI for $OS $VERSION"
	case "$PACKAGETYPE" in
		apt)
			# Ideally we want to use curl, but on some installs we
			# only have wget. Detect and use what's available.
			CURL=
			if type curl >/dev/null; then
				CURL="curl -fsSL"
			elif type wget >/dev/null; then
				CURL="wget -q -O-"
			fi
			if [ -z "$CURL" ]; then
				echo "The installer needs either curl or wget to download files."
				echo "Please install either curl or wget to proceed."
				exit 1
			fi
			export DEBIAN_FRONTEND=noninteractive
			set -x
            case $ARCH in
                aarch64|arm64)
                    #
                    ;;
                386)
                    #
                    ;;
                arm)
                #
                    ;;
            esac
			# $SUDO mkdir -p --mode=0755 /usr/share/keyrings
			# case "$APT_KEY_TYPE" in
			# 	legacy)
			# 		$CURL "https://pkgs.tailscale.com/$TRACK/$OS/$VERSION.asc" | $SUDO apt-key add -
			# 		$CURL "https://pkgs.tailscale.com/$TRACK/$OS/$VERSION.list" | $SUDO tee /etc/apt/sources.list.d/tailscale.list
			# 	;;
			# 	keyring)
			# 		$CURL "https://pkgs.tailscale.com/$TRACK/$OS/$VERSION.noarmor.gpg" | $SUDO tee /usr/share/keyrings/tailscale-archive-keyring.gpg >/dev/null
			# 		$CURL "https://pkgs.tailscale.com/$TRACK/$OS/$VERSION.tailscale-keyring.list" | $SUDO tee /etc/apt/sources.list.d/tailscale.list
			# 	;;
			# esac
			# $SUDO apt-get update
			# $SUDO apt-get install -y tailscale
			# if [ "$APT_SYSTEMCTL_START" = "true" ]; then
			# 	$SUDO systemctl enable --now tailscaled
			# 	$SUDO systemctl start tailscaled
			# fi
			set +x
		;;
		yum)
			set -x
			$SUDO yum install yum-utils -y
			$SUDO yum-config-manager -y --add-repo "https://pkgs.tailscale.com/$TRACK/$OS/$VERSION/tailscale.repo"
			$SUDO yum install tailscale -y
			$SUDO systemctl enable --now tailscaled
			set +x
		;;
		dnf)
			set -x
			$SUDO dnf config-manager --add-repo "https://pkgs.tailscale.com/$TRACK/$OS/$VERSION/tailscale.repo"
			$SUDO dnf install -y tailscale
			$SUDO systemctl enable --now tailscaled
			set +x
		;;
		zypper)
			set -x
			$SUDO zypper ar -g -r "https://pkgs.tailscale.com/$TRACK/$OS/$VERSION/tailscale.repo"
			$SUDO zypper ref
			$SUDO zypper in tailscale
			$SUDO systemctl enable --now tailscaled
			set +x
			;;
		pacman)
			set -x
			$SUDO pacman -S tailscale --noconfirm
			$SUDO systemctl enable --now tailscaled
			set +x
			;;
		pkg)
			set -x
			$SUDO pkg install tailscale
			$SUDO service tailscaled enable
			$SUDO service tailscaled start
			set +x
			;;
		apk)
			set -x
			$SUDO apk add tailscale
			$SUDO rc-update add tailscale
			set +x
			;;
		xbps)
			set -x
			$SUDO xbps-install tailscale -y 
			set +x
			;;
		emerge)
			set -x
			$SUDO emerge --ask=n net-vpn/tailscale
			set +x
			;;
		appstore)
			set -x
			open "https://apps.apple.com/us/app/tailscale/id1475387142"
			set +x
			;;
		*)
			echo "unexpected: unknown package type $PACKAGETYPE"
			exit 1
			;;
	esac

	echo "Installation complete! Log in to start using Tailscale by running:"
	echo
	if [ -z "$SUDO" ]; then
		echo "tailscale up"
	else
		echo "$SUDO tailscale up"
	fi
}

main

}

