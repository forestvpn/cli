#!/bin/bash

curl https://github.com/forestvpn/cli/releases/download/latest/forest -o /usr/local/bin/forest
curl https://github.com/forestvpn/cli/releases/download/latest/forestd -o /usr/local/bin/forestd
curl https://raw.githubusercontent.com/forestvpn/cli/main/forestd/env/forestd.service -o /etc/systemd/system/forestd.service
systemctl daemon-reload
systemctl start forestd
forest --help