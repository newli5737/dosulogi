#!/usr/bin/env bash
# Generate JWT secrets and append to .env (or print for copy-paste)
set -euo pipefail

ACCESS=$(openssl rand -hex 32)
REFRESH=$(openssl rand -hex 32)

cat <<EOF
# Paste into .env:
JWT_ACCESS_SECRET=${ACCESS}
JWT_REFRESH_SECRET=${REFRESH}
JWT_ACCESS_TTL_MIN=15
JWT_REFRESH_TTL_DAY=7
JWT_ADMIN_REFRESH_TTL_DAY=14
COOKIE_DOMAIN=.dosutech.site
EOF
