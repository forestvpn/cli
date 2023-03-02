#!/bin/sh
set -eu

GITHUB_TOKEN=${GITHUB_TOKEN:-""}

echo "machine github.com login token password ${GITHUB_TOKEN}" >> ~/.netrc
git config --global url."https://${GITHUB_TOKEN}@github.com/forestvpn/".insteadOf "https://github.com/forestvpn/"
