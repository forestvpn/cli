#!/bin/sh
ARCHS="arm64"
ARCHS="$ARCHS mips"
OUT_DIR="$DRONE_WORKSPACE/packages/ipkg"
SOURCE_DIR="$OUT_DIR/src"

for arch in ${ARCHS}; do
    ln -s "$SOURCE_DIR/debian-binary" "$SOURCE_DIR/$arch/debian-binary"
    mkdir -p "$SOURCE_DIR/$arch/control"
    ln -s "$SOURCE_DIR/postinst" "$SOURCE_DIR/$arch/control/postinst"
    CONTROL_FILE="$SOURCE_DIR/$arch/control/control"
    touch $CONTROL_FILE
    echo "Package: fvpn" >> $CONTROL_FILE
    echo "Version: ${DRONE_TAG}" >> $CONTROL_FILE
    echo "Architecture: $arch" >> $CONTROL_FILE
    echo "Maintainer: nikiforova693@gmail.com" >> $CONTROL_FILE
    echo "Description: ForestVPN - Fast, secure, and modern VPN." >> $CONTROL_FILE
    echo "Priority: optional" >> $CONTROL_FILE
    echo "Depends: wireguard" >> $CONTROL_FILE
    BIN_DIR="$SOURCE_DIR/$arch/data/usr/local/bin/"
    mkdir -p $BIN_DIR
    ln -s $DRONE_WORKSPACE/src/dist/fvpn-linux-$arch/fvpn $BIN_DIR/fvpn
    pushd $SOURCE_DIR/$arch/control
    tar --numeric-owner --group=0 --owner=0 -czf /tmp/control.tar.gz ./*
    popd
    pushd $SOURCE_DIR/$arch/data
    tar --numeric-owner --group=0 --owner=0 -czf /tmp/data.tar.gz ./*
    popd
    pushd $SOURCE_DIR/$arch
    tar --numeric-owner --group=0 --owner=0 -cf $OUT_DIR/fvpn_$DRONE_TAG.$arch.ipk $SOURCE_DIR/$arch/debian-binary /tmp/data.tar.gz /tmp/control.tar.gz 
    popd
done