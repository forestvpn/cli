#!/bin/bash
ARCHS="arm64 mips"
SOURCE_DIR="$OUT_DIR/src"

for arch in $ARCHS; do
    cp "debian-binary" "./$arch/"
    mkdir -p "./$arch/control"
    cp "postinst" "./$arch/control/"
    CONTROL_FILE="./$arch/control/control"
    touch $CONTROL_FILE
    echo "Package: fvpn" >> $CONTROL_FILE
    echo "Version: ${DRONE_TAG}" >> $CONTROL_FILE
    echo "Architecture: $arch" >> $CONTROL_FILE
    echo "Maintainer: nikiforova693@gmail.com" >> $CONTROL_FILE
    echo "Description: ForestVPN - Fast, secure, and modern VPN." >> $CONTROL_FILE
    echo "Priority: optional" >> $CONTROL_FILE
    echo "Depends: wireguard" >> $CONTROL_FILE
    BIN_DIR="./$arch/data/usr/local/bin/"
    mkdir -p $BIN_DIR

    if [$arch = "mips"]; then
        cp ../../src/dist/fvpn-linux-mips_hardfloat/fvpn $BIN_DIR
    else
        cp ../../src/dist/fvpn-linux-$arch/fvpn $BIN_DIR
    fi
    pushd ./$arch/control
    tar --numeric-owner --group=0 --owner=0 -czf ./$arch/control.tar.gz ./*
    popd
    pushd ./$arch/data
    tar --numeric-owner --group=0 --owner=0 -czf ./$arch/data.tar.gz ./*
    popd
    pushd ./$arch
    tar --numeric-owner --group=0 --owner=0 -cf ../fvpn_$DRONE_TAG.$arch.ipk ./$arch/debian-binary ./$arch/data.tar.gz ./$arch/control.tar.gz 
    popd
done