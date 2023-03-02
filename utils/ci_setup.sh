#!/bin/sh
set -eu

GITHUB_TOKEN=${GITHUB_TOKEN:-""}

git config --global url."https://token:${GITHUB_TOKEN}@github.com/forestvpn/".insteadOf "https://github.com/forestvpn/"
