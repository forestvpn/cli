#!/bin/bash
ARCHS="arm64 mips"
SOURCE_DIR="$OUT_DIR/src"

for arch in $ARCHS; do
    cp "/drone/src/ipkg/src/debian-binary" "/drone/src/ipkg/src/$arch/debian-binary"
    mkdir -p "/drone/src/ipkg/src/$arch/control"
    cp "/drone/src/ipkg/src/postinst" "/drone/src/ipkg/src/$arch/control/postinst"
    CONTROL_FILE="/drone/src/ipkg/src/$arch/control/control"
    touch $CONTROL_FILE
    echo "Package: fvpn" >> $CONTROL_FILE
    echo "Version: ${DRONE_TAG}" >> $CONTROL_FILE
    echo "Architecture: $arch" >> $CONTROL_FILE
    echo "Maintainer: nikiforova693@gmail.com" >> $CONTROL_FILE
    echo "Description: ForestVPN - Fast, secure, and modern VPN." >> $CONTROL_FILE
    echo "Priority: optional" >> $CONTROL_FILE
    echo "Depends: wireguard" >> $CONTROL_FILE
    BIN_DIR="/drone/src/ipkg/src/$arch/data/usr/local/bin/"
    mkdir -p $BIN_DIR

    if [[ "$arch" == "mips" ]]; then
        cp /drone/src/src/dist/fvpn_linux_mips_hardfloat/fvpn $BIN_DIR
    else
        cp /drone/src/src/dist/fvpn_linux_$arch/fvpn $BIN_DIR
    fi
    pushd /drone/src/ipkg/src/$arch/control
    tar --numeric-owner --group=0 --owner=0 -czf ./$arch/control.tar.gz ./*
    popd
    pushd /drone/src/ipkg/src/$arch/data
    tar --numeric-owner --group=0 --owner=0 -czf ./$arch/data.tar.gz ./*
    popd
    pushd /drone/src/ipkg/src/$arch
    tar --numeric-owner --group=0 --owner=0 -cf /drone/src/ipkg/fvpn_$DRONE_TAG.$arch.ipk /drone/src/ipkg/src/$arch/debian-binary /drone/src/ipkg/src/$arch/data.tar.gz /drone/src/ipkg/src/$arch/control.tar.gz 
    popd
done