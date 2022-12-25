# !/bin/sh
export GOOS=linux
export GOARCH=mips
cd src
go build -ldflags="-w -s -X main.appVersion=${VERSION}" -gcflags="-l -B -wb=false" -o ../fvpn
cd ../
apt update
apt install binutils -y
mkdir -p ipkg/fvpn/control
mkdir ipkg/fvpn/data
export CTRLF=ipkg/fvpn/control/control
echo Package: fvpn > $CTRLF
echo Version: ${VERSION} >> $CTRLF
echo Architecture: $GOARCH >> $CTRLF
echo Maintainer: nikiforova693@gmail.com >> $CTRLF
echo Description: ForestVPN - Fast, secure, and modern VPN. >> $CTRLF
echo Priority: optional >> $CTRLF
echo Depends: wireguard >> $CTRLF
echo 2.0 > ipkg/fvpn/debian-binary
export PREINST=ipkg/fvpn/control/preinst
echo "#!/bin/bash" > $PREINST
echo arch=`uname -m` > $PREINST
echo "if [ $arch -ne '$GOARCH' ]; then" >> $PREINST
echo "    echo ${arch} is not supported" >> $PREINST
echo "    exit 1" >> $PREINST
chmod +x $PREINST
export POSTINST=ipkg/fvpn/control/postinst
echo "#!/bin/bash" > $POSTINST
echo fvpn >> $POSTINST
echo "if [ $? -eq 0 ]; then" >> $POSTINST
echo "    echo Successfully installed" >> $POSTINST
echo else >> $POSTINST
echo "    echo Could not install >&2" >> $POSTINST
chmod +x ipkg/fvpn/control/postinst
export BIN_DIR=ipkg/fvpn/data/usr/local/bin/
mkdir -p $BIN_DIR
cp fvpn $BIN_DIR
./pack.bash
rm fvpn
rm -rf ipkg