#!/usr/bin/env bash
set -euo pipefail

APP_DIR="/home/dosulogi"
REPO_URL="${REPO_URL:-https://github.com/newli5737/dosulogi.git}"
GO_VERSION="${GO_VERSION:-1.23.4}"

install_go() {
  if /usr/local/go/bin/go version 2>/dev/null | grep -q "go1.2"; then
    return
  fi
  echo "==> Install Go ${GO_VERSION}"
  tmp="$(mktemp -d)"
  curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -o "$tmp/go.tgz"
  rm -rf /usr/local/go
  tar -C /usr/local -xzf "$tmp/go.tgz"
  export PATH="/usr/local/go/bin:$PATH"
}

if [[ "$(id -u)" -eq 0 ]]; then
  SUDO=""
else
  SUDO="sudo"
fi

if [[ ! -d "$APP_DIR/.git" ]]; then
  echo "==> Clone repo to $APP_DIR"
  git clone "$REPO_URL" "$APP_DIR"
fi

cd "$APP_DIR"
git pull origin main

install_go
export PATH="/usr/local/go/bin:${PATH:-}"

echo "==> Build backend"
go build -o server ./cmd/server

echo "==> Build frontend"
if [[ -d frontend ]]; then
  cd frontend
  npm ci
  npm run build
  cd "$APP_DIR"
fi

echo "==> Prepare dirs"
mkdir -p "$APP_DIR/uploads/contracts" "$APP_DIR/uploads/invoices" "$APP_DIR/frontend/dist"
if [[ ! -f "$APP_DIR/.env" ]]; then
  cp "$APP_DIR/.env.example" "$APP_DIR/.env"
  echo "!! Created $APP_DIR/.env from example — update secrets before production"
fi

echo "==> Nginx"
$SUDO cp deploy/nginx/api-logi.dosutech.site.conf /etc/nginx/sites-available/
$SUDO cp deploy/nginx/logi.dosutech.site.conf /etc/nginx/sites-available/
$SUDO ln -sf /etc/nginx/sites-available/api-logi.dosutech.site.conf /etc/nginx/sites-enabled/
$SUDO ln -sf /etc/nginx/sites-available/logi.dosutech.site.conf /etc/nginx/sites-enabled/
$SUDO nginx -t
$SUDO systemctl reload nginx

echo "==> Systemd"
$SUDO cp deploy/systemd/dosulogi.service /etc/systemd/system/
$SUDO systemctl daemon-reload
$SUDO systemctl enable dosulogi
$SUDO systemctl restart dosulogi

echo "Done."
echo "  API: http://api-logi.dosutech.site"
echo "  FE:  http://logi.dosutech.site"
echo "  SSL: certbot --nginx -d api-logi.dosutech.site -d logi.dosutech.site"
