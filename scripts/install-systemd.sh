#!/usr/bin/env bash
set -euo pipefail

APP_DIR="${APP_DIR:-/srv/vue-nav-page}"
PORT="${PORT:-21000}"
HOST="${HOST:-0.0.0.0}"
SERVICE_NAME="${SERVICE_NAME:-vue-nav-page.service}"

if ! command -v go >/dev/null 2>&1; then
  echo "Go is required. Install Go 1.22+ first." >&2
  exit 1
fi

mkdir -p "$APP_DIR/assets"
cp index.html config.json main.go go.mod "$APP_DIR/"
cp assets/vue.global.prod.js "$APP_DIR/assets/"

(
  cd "$APP_DIR"
  go build -trimpath -ldflags='-s -w' -o nav-server main.go
)

cat > "/etc/systemd/system/$SERVICE_NAME" <<EOF_SERVICE
[Unit]
Description=Vue Nav Page static HTTP server
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
WorkingDirectory=$APP_DIR
ExecStart=$APP_DIR/nav-server -host $HOST -port $PORT
Restart=on-failure
RestartSec=3
User=root

[Install]
WantedBy=multi-user.target
EOF_SERVICE

systemctl daemon-reload
systemctl enable --now "$SERVICE_NAME"
systemctl status "$SERVICE_NAME" --no-pager --plain
