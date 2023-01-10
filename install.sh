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
			#OS_UNSUPPORTED=1
			;;
		*)
			#OS_UNSUPPORTED=1
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
	case "$PACKAGETYPE" in
		apt)
			export DEBIAN_FRONTEND=noninteractive
			set -x
            case $ARCH in
                aarch64|arm64)
                    $CURL "https://github.com/forestvpn/cli/releases/latest/download/fvpn_linux_arm64.deb" > fvpn_linux_arm64.deb
					$SUDO dpkg -i fvpn_linux_arm64.deb
					if [ $? -eq 1 ] 
					then 
						apt install -fy
					else 
						rm fvpn_linux_arm64.deb
					fi
                    ;;
                arm)
                    $CURL "https://github.com/forestvpn/cli/releases/latest/download/fvpn_linux_arm.deb" > fvpn_linux_arm.deb
					$SUDO dpkg -i fvpn_linux_arm.deb
					if [ $? -eq 1 ] 
					then 
						apt install -fy
					else 
						rm fvpn_linux_arm.deb
					fi
                    ;;
				amd64)
					$CURL "https://github.com/forestvpn/cli/releases/latest/download/fvpn_linux_amd64.deb" > fvpn_linux_amd64.deb
					$SUDO dpkg -i fvpn_linux_amd64.deb
					if [ $? -eq 1 ] 
					then 
						apt install -fy
					else 
						rm fvpn_linux_amd64.deb
					fi
                    ;;				
                386)
                    $CURL "https://github.com/forestvpn/cli/releases/latest/download/fvpn_linux_386.deb" > fvpn_linux_386.deb
					$SUDO dpkg -i fvpn_linux_386.deb
					if [ $? -eq 1 ] 
					then 
						apt install -fy
					else 
						rm fvpn_linux_386.deb
					fi
                    ;;
            esac
			set +x
			;;
		yum)
			set -x
			case $ARCH in
				aarch64|arm64)
                    $CURL "https://github.com/forestvpn/cli/releases/latest/download/fvpn_linux_arm64.rpm" > fvpn_linux_arm64.rpm
					$SUDO yum localinstall fvpn_linux_arm64.rpm
					if [ $? -eq 0 ] 
					then
						rm fvpn_linux_arm64.rpm
					fi
                    ;;
                arm)
                    $CURL "https://github.com/forestvpn/cli/releases/latest/download/fvpn_linux_arm.rpm" > fvpn_linux_arm.rpm
					$SUDO yum localinstall -i fvpn_linux_arm.rpm
					if [ $? -eq 0 ] 
					then
						rm fvpn_linux_arm.rpm
					fi
                    ;;
				amd64)
					$CURL "https://github.com/forestvpn/cli/releases/latest/download/fvpn_linux_amd64.rpm" > fvpn_linux_amd64.rpm
					$SUDO yum localinstall fvpn_linux_amd64.rpm
					if [ $? -eq 0 ] 
					then
						rm fvpn_linux_amd64.rpm
					fi
					;;
                386)
                    $CURL "https://github.com/forestvpn/cli/releases/latest/download/fvpn_linux_386.rpm" > fvpn_linux_386.rpm
					$SUDO yum localinstall -i fvpn_linux_386.rpm
					if [ $? -eq 0 ] 
					then
						rm fvpn_linux_386.rpm
					fi
                    ;;
			esac			
			set +x
			;;
		dnf)
			set -x
			case $ARCH in
				aarch64|arm64)
                    $CURL "https://github.com/forestvpn/cli/releases/latest/download/fvpn_linux_arm64.rpm" > fvpn_linux_arm64.rpm
					$SUDO dnf localinstall fvpn_linux_arm64.rpm
					if [ $? -eq 0 ] 
					then
						rm fvpn_linux_arm64.rpm
					fi
                    ;;
                arm)
                    $CURL "https://github.com/forestvpn/cli/releases/latest/download/fvpn_linux_arm.rpm" > fvpn_linux_arm.rpm
					$SUDO dnf localinstall -i fvpn_linux_arm.rpm
					if [ $? -eq 0 ] 
					then
						rm fvpn_linux_arm.rpm
					fi
                    ;;
				amd64)
					$CURL "https://github.com/forestvpn/cli/releases/latest/download/fvpn_linux_amd64.rpm" > fvpn_linux_amd64.rpm
					$SUDO dnf localinstall fvpn_linux_amd64.rpm
					if [ $? -eq 0 ] 
					then
						rm fvpn_linux_amd64.rpm
					fi
					;;
                386)
                    $CURL "https://github.com/forestvpn/cli/releases/latest/download/fvpn_linux_386.rpm" > fvpn_linux_386.rpm
					$SUDO dnf localinstall -i fvpn_linux_386.rpm
					if [ $? -eq 0 ] 
					then
						rm fvpn_linux_386.rpm
					fi
                    ;;
			esac
			set +x
			;;
		apk)
			set -x
			case $ARCH in
				aarch64|arm64)
                    $CURL "https://github.com/forestvpn/cli/releases/latest/download/fvpn_linux_arm64.apk" > fvpn_linux_arm64.apk
					$SUDO apk add fvpn_linux_arm64.apk --allow-untrusted
					if [ $? -eq 0 ] 
					then
						rm fvpn_linux_arm64.apk
					fi
                    ;;
                arm)
                    $CURL "https://github.com/forestvpn/cli/releases/latest/download/fvpn_linux_arm.apk" > fvpn_linux_arm.apk
					$SUDO apk add -i fvpn_linux_arm.apk --allow-untrusted
					if [ $? -eq 0 ] 
					then
						rm fvpn_linux_arm.apk
					fi
                    ;;
				amd64)
					$CURL "https://github.com/forestvpn/cli/releases/latest/download/fvpn_linux_amd64.apk" > fvpn_linux_amd64.apk
					$SUDO apk add fvpn_linux_amd64.apk --allow-untrusted
					if [ $? -eq 0 ] 
					then
						rm fvpn_linux_amd64.apk
					fi
					;;
                386)
                    $CURL "https://github.com/forestvpn/cli/releases/latest/download/fvpn_linux_386.apk" > fvpn_linux_386.apk
					$SUDO apk add -i fvpn_linux_386.apk --allow-untrusted
					if [ $? -eq 0 ] 
					then
						rm fvpn_linux_386.apk
					fi
                    ;;
			esac
			set +x
			;;
		brew)
			brew install forestvpn/stable/fvpn
			;;
		*)
			echo "unexpected: unknown package type $PACKAGETYPE"
			exit 1
			;;
	esac

	echo "Installation complete! Log in to start using ForestVPN by running:"
	echo fvpn account login
	echo ""
}

main

}

