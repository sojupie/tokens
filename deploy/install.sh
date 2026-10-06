#!/bin/sh
set -eu

if [ "$(id -u)" != 0 ]; then
    echo 'Run with sudo: sudo sh deploy/install.sh' >&2
    exit 1
fi
if [ ! -f /opt/tokens/deploy/compose.yaml ]; then
    echo 'Place the checkout at /opt/tokens before installing.' >&2
    exit 1
fi
docker compose version >/dev/null
docker compose -f /opt/tokens/deploy/compose.yaml config --quiet
update_state=$(systemctl show tokens-update.service --property=ActiveState --value 2>/dev/null || true)
case "$update_state" in
    active|activating|deactivating)
        echo 'Wait for the running update to finish before installing units.' >&2
        exit 1
        ;;
esac

if [ -d /var/lib/tokens ] && [ "$(stat -c %u /var/lib/tokens)" != 10001 ]; then
    echo '/var/lib/tokens must belong to UID 10001. Review ownership before migrating existing data.' >&2
    exit 1
fi
install -d -o 10001 -g 10001 -m 0700 /var/lib/tokens
install -d -m 0755 /etc/tokens
if [ ! -e /etc/tokens/cloudflared.env ]; then
    install -m 0600 /opt/tokens/deploy/cloudflared.env.example /etc/tokens/cloudflared.env
fi
if [ ! -e /etc/tokens/update.env ]; then
    install -m 0600 /opt/tokens/deploy/systemd/update.env.example /etc/tokens/update.env
fi
chown root:root /etc/tokens/cloudflared.env /etc/tokens/update.env
chmod 0600 /etc/tokens/cloudflared.env /etc/tokens/update.env
install -m 0644 /opt/tokens/deploy/systemd/tokens.service \
    /opt/tokens/deploy/systemd/tokens-update.service \
    /opt/tokens/deploy/systemd/tokens-update-failure.service \
    /opt/tokens/deploy/systemd/tokens-update.timer /etc/systemd/system/
systemctl daemon-reload
echo 'Installed. Set the tunnel token, build the images, then create and approve the baseline before enabling services.'
