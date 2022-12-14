#!/bin/sh
ARCHS=("arm64", "mips")
PKG_DIR="packages/ipkg/"

for arch in ${ARCHS[@]}; do
    CONTROL_FILE_PATH="$PKG_DIR/$arch/control/control"
    BIN_DIR="$PKG_DIR/$arch/data/usr/local/bin/"
    mkdir -p $PKG_DIR/$arch/{control,data}
    touch CONTROL_FILE_PATH
    echo "Package: fvpn" >> $CONTROL_FILE_PATH
    echo "Version: ${DRONE_TAG}" >> $CONTROL_FILE_PATH
    echo "Architecture: $arch" >> $CONTROL_FILE_PATH
    echo "Maintainer: nikiforova693@gmail.com" >> $CONTROL_FILE_PATH
    echo "Description: ForestVPN - Fast, secure, and modern VPN." >> $CONTROL_FILE_PATH
    echo "Priority: optional" >> $CONTROL_FILE_PATH
    echo "Depends: wireguard" >> $CONTROL_FILE_PATH
    echo "2.0" > "$PKG_DIR/debian-binary"
    mkdir -p $BIN_DIR
    ln -s ../src/dist/fvpn-linux-$arch/fvpn $BIN_DIR/fvpn
    pushd $PKG_DIR/control
    tar --numeric-owner --group=0 --owner=0 -czf /tmp/control.tar.gz ./*
    popd
    pushd $PKG_DIR/data
    tar --numeric-owner --group=0 --owner=0 -czf /tmp/data.tar.gz ./*
    popd
    pushd $PKG_DIR
    tar --numeric-owner --group=0 --owner=0 -cf fvpn_$DRONE_TAG.$arch.ipk $PKG_DIR/debian-binary /tmp/data.tar.gz /tmp/control.tar.gz 
    popd
    rm $BIN_DIR/fvpn
done