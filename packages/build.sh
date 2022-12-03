#!/bin/sh
PKG_DIR="packages/ipkg/fvpn"
CONTROL_FILE_PATH="$PKG_DIR/control/control"
BIN_DIR="$PKG_DIR/data/usr/local/bin/"
mkdir -p $PKG_DIR/{control,data}
touch CONTROL_FILE_PATH
echo "Package: fvpn" >> $CONTROL_FILE_PATH
echo "Version: ${DRONE_TAG}" >> $CONTROL_FILE_PATH
echo "Architecture: arm64" >> $CONTROL_FILE_PATH
echo "Maintainer: nikiforova693@gmail.com" >> $CONTROL_FILE_PATH
echo "Description: ForestVPN - Fast, secure, and modern VPN." >> $CONTROL_FILE_PATH
echo "Priority: optional" >> $CONTROL_FILE_PATH
echo "Depends: wireguard" >> $CONTROL_FILE_PATH
echo "2.0" > "$PKG_DIR/debian-binary"
mkdir -p $BIN_DIR
ln -s ../src/dist/fvpn-linux-arm64/fvpn $BIN_DIR/fvpn
pushd $PKG_DIR/control
tar --numeric-owner --group=0 --owner=0 -czf /tmp/control.tar.gz ./*
popd
pushd $PKG_DIR/data
tar --numeric-owner --group=0 --owner=0 -czf /tmp/data.tar.gz ./*
popd
pushd $PKG_DIR
tar --numeric-owner --group=0 --owner=0 -cf fvpn_$DRONE_TAG.arm64.ipk $PKG_DIR/debian-binary /tmp/data.tar.gz /tmp/control.tar.gz 
popd
rm $BIN_DIR/fvpn