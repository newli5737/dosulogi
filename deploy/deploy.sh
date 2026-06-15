#!/usr/bin/env bash
set -euo pipefail

APP_DIR="/var/www/dosulogi"
REPO_DIR="$(cd "$(dirname "$0")/.." && pwd)"

echo "==> Build backend"
cd "$REPO_DIR"
GOOS=linux GOARCH=amd64 go build -o server ./cmd/server

echo "==> Build frontend"
if [ -d "$REPO_DIR/frontend" ]; then
  cd "$REPO_DIR/frontend"
  npm ci
  npm run build
fi

echo "==> Deploy to $APP_DIR"
sudo mkdir -p "$APP_DIR"/{uploads/contracts,uploads/invoices,frontend/dist}
sudo cp "$REPO_DIR/server" "$APP_DIR/server"
sudo cp "$REPO_DIR/.env" "$APP_DIR/.env" 2>/dev/null || sudo cp "$REPO_DIR/.env.example" "$APP_DIR/.env"
if [ -d "$REPO_DIR/frontend/dist" ]; then
  sudo rsync -a --delete "$REPO_DIR/frontend/dist/" "$APP_DIR/frontend/dist/"
fi

echo "==> Install nginx configs"
sudo cp "$REPO_DIR/deploy/nginx/api-logi.dosutech.site.conf" /etc/nginx/sites-available/
sudo cp "$REPO_DIR/deploy/nginx/logi.dosutech.site.conf" /etc/nginx/sites-available/
sudo ln -sf /etc/nginx/sites-available/api-logi.dosutech.site.conf /etc/nginx/sites-enabled/
sudo ln -sf /etc/nginx/sites-available/logi.dosutech.site.conf /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx

echo "==> Restart API service"
sudo cp "$REPO_DIR/deploy/systemd/dosulogi.service" /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable dosulogi
sudo systemctl restart dosulogi

echo "Done."
echo "  API: http://api-logi.dosutech.site"
echo "  FE:  http://logi.dosutech.site"
