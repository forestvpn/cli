#!/bin/bash

curl -s https://github.com/forestvpn/cli/releases/download/latest/forest -o /tmp/forest
mv /tmp/forest /usr/local/bin/
chmod +x /usr/local/bin/forest
curl -s https://github.com/forestvpn/cli/releases/download/latest/forestd -o /tmp/forestd
mv /tmp/forestd /usr/local/bin/
chmod +x /usr/local/bin/forestd
curl -s https://raw.githubusercontent.com/forestvpn/cli/main/forestd/env/forestd.service -o /etc/systemd/system/forestd.service
systemctl daemon-reload
systemctl start forestd
forest --help