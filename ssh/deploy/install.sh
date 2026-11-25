#!/usr/bin/env bash
# Run on the server as root, from the directory that holds portfolio-ssh-linux
# and portfolio-ssh.service.  Your own admin sshd MUST already be off port 22
# (e.g. Port 2222 in /etc/ssh/sshd_config, firewall open, tested from a second
# terminal) or this will fail to bind and you can lock yourself out.
set -euo pipefail
id portfolio >/dev/null 2>&1 || useradd --system --home /opt/portfolio-ssh --shell /usr/sbin/nologin portfolio
install -d -o portfolio -g portfolio /opt/portfolio-ssh
install -m 0755 -o portfolio -g portfolio portfolio-ssh-linux /opt/portfolio-ssh/portfolio-ssh
install -m 0644 portfolio-ssh.service /etc/systemd/system/portfolio-ssh.service
systemctl daemon-reload
systemctl enable --now portfolio-ssh
systemctl --no-pager status portfolio-ssh | head -8
