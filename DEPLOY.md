# Deploying the netlink backend

The backend is a single Go binary built from `backend/Dockerfile` (multi-stage,
distroless-ish alpine). Any host that builds a Dockerfile works — Koyeb, Hugging
Face Spaces (Docker), Railway, Render, Fly, etc.

## What the host needs
- **Build context / root:** `backend/`
- **Dockerfile:** `backend/Dockerfile`
- **Port:** the app reads `$PORT` (defaults to 8080). Most hosts inject `PORT`
  automatically — don't hard-set it.
- **Health check:** `GET /api/health` → `200`.

## Required environment variables
Copy the values from `backend/.env` (git-ignored) into the host's env settings:

| Var | Purpose |
| --- | --- |
| `DATABASE_URL` | Postgres (Neon). Schema is auto-created on first boot via `InitSchema()`. |
| `SUPABASE_URL` / `SUPABASE_SERVICE_KEY` / `SUPABASE_BUCKET` | file/voice storage |
| `ALLOWED_ORIGINS` | comma-separated frontend origins (CORS + WS origin check) |
| `APP_URL` | public frontend URL (used in emails) |
| `RESEND_API_KEY` *(or `GMAIL_USERNAME` + `GMAIL_APP_PASSWORD` + `FROM_EMAIL`)* | transactional email |
| `ENV` | `production` |

No migration step is required — the app creates its tables on startup.

## Keep-alive (free tiers that sleep)
`.github/workflows/keepalive.yml` pings `/api/health` every ~12 min. Set the repo
secret `RENDER_URL` (or edit it) to your deployed base URL. For rock-solid
pinging use cron-job.org against `<base-url>/api/health` instead.

## Render (if you add a card)
`render.yaml` is a ready Blueprint — Render Dashboard → New → Blueprint → this repo.
Note: Render now requires a payment card on file even for free instances.
