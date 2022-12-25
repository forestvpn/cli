#!/bin/bash
pushd ipkg/fvpn/control/
tar --numeric-owner --group=0 --owner=0 -czf ../control.tar.gz ./*
popd
pushd ipkg/fvpn/data
tar --numeric-owner --group=0 --owner=0 -czf ../data.tar.gz ./*
popd
pushd ipkg/fvpn
tar --numeric-owner --group=0 --owner=0 -cf ../../fvpn_${VERSION}_${GOARCH}.ipk ./debian-binary ./data.tar.gz ./control.tar.gz
popd