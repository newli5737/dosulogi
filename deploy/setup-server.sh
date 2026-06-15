#!/usr/bin/env bash
set -euo pipefail

cd /home/dosulogi
git pull origin main

cat > .env <<'EOF'
APP_PORT=8089
APP_ENV=production
DB_HOST=127.0.0.1
DB_PORT=5432
DB_NAME=dosulogi
DB_USER=postgres
DB_PASSWORD=test1234
REDIS_ADDR=127.0.0.1:6379
REDIS_PASSWORD=
JWT_ACCESS_SECRET=prod-access-secret-dosulogi-256bit-key
JWT_REFRESH_SECRET=prod-refresh-secret-dosulogi-256bit-key
JWT_ACCESS_TTL_MIN=15
JWT_REFRESH_TTL_DAY=7
SEPAY_WEBHOOK_SECRET=dev-sepay-secret
FROM_EMAIL=no-reply@dosulogi.com
TRACKING_API_BASE_URL=https://api.trackingprovider.com
TRACKING_API_KEY=
TRACKING_POLL_INTERVAL_SEC=300
TRACKING_WEBHOOK_SECRET=dev-tracking-secret
UPLOAD_DIR=/home/dosulogi/uploads
CORS_ORIGINS=https://logi.dosutech.site,http://logi.dosutech.site
EOF

sudo -u postgres psql -tc "SELECT 1 FROM pg_database WHERE datname='dosulogi'" | grep -q 1 || sudo -u postgres createdb dosulogi
sudo -u postgres psql -c "ALTER USER postgres WITH PASSWORD 'test1234';" || true

chmod +x deploy/deploy.sh
bash deploy/deploy.sh

echo "==> Certbot SSL (skip if cert exists)"
if [ ! -f /etc/letsencrypt/live/api-logi.dosutech.site/fullchain.pem ]; then
  certbot --nginx -d api-logi.dosutech.site -d logi.dosutech.site --non-interactive --agree-tos -m admin@dosutech.site --redirect
else
  echo "SSL cert already present"
fi

systemctl status dosulogi --no-pager || true
curl -s http://127.0.0.1:8089/health || true
