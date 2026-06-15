# dosulogi

Logistics ERP/CRM — backend Go (Gin) + frontend React (Vite).

## Domains (production)

| Service | Domain | Port |
|---------|--------|------|
| API | http://api-logi.dosutech.site | 8089 (internal) |
| Frontend | http://logi.dosutech.site | 80 (nginx static) |

## Local development

```bash
# Backend
cp .env.example .env
go run ./cmd/server

# Frontend
cd frontend
npm install
npm run dev
```

- API: http://127.0.0.1:8089
- FE: http://localhost:5173 (proxy `/api` → backend)

## Default admin

- Email: `admin@dosulogi.com`
- Password: `Admin@123`

## Deploy (VPS)

```bash
chmod +x deploy/deploy.sh
./deploy/deploy.sh
```

Nginx configs:
- `deploy/nginx/api-logi.dosutech.site.conf` — reverse proxy → `:8089`
- `deploy/nginx/logi.dosutech.site.conf` — serve `frontend/dist`

Systemd: `deploy/systemd/dosulogi.service`

## Stack

- Go 1.22+ · Gin · PostgreSQL · Redis (optional)
- JWT auth + RBAC
- SePay webhook · Tracking API (3rd party)
- Email: log-only mailer (no SendGrid)
