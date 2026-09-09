# Environment Variables

Three separate places hold environment configuration, because `apps/server` and `apps/web` are
deployed independently:

| File | Used by | Committed? |
|---|---|---|
| `apps/server/.env` | Go server, running natively (`make dev`/`make run`) | No — gitignored |
| `apps/server/.env.example` | Reference for the above | Yes |
| `apps/web/.env.local` | Next.js dev server | No — gitignored |
| `apps/web/.env.example` | Reference for the above | Yes |
| `infra/.env` | Docker Compose (mongo + server + seed) | No — gitignored |
| `infra/.env.example` | Reference for the above | Yes |

**Never commit `.env`, `.env.local`, or any file with real secrets.** Only `.env.example` files are
committed, with empty or placeholder values.

---

## `apps/server` — Go API

| Variable | Required | Default | Description |
|---|---|---|---|
| `MONGO_URI` | Yes | — | MongoDB connection string |
| `DB_NAME` | No | `reliva` | Database name |
| `JWT_SECRET` | Yes | — | Signs both `pending` and `access` JWTs |
| `SERVER_PORT` | No | `8080` | HTTP listen port |
| `ALLOWED_ORIGINS` | No | `http://localhost:3000` | Comma-separated CORS allowlist |
| `CRON_SECRET` | No | `""` | Bearer secret protecting `POST /api/v1/cron/notify` |
| `VAPID_PUBLIC_KEY` | No | `""` | Web Push — Phase 5 |
| `VAPID_PRIVATE_KEY` | No | `""` | Web Push — Phase 5 |
| `VAPID_SUBJECT` | No | `""` | Web Push contact, e.g. `mailto:you@example.com` |
| `GIN_MODE` | No | debug if unset | Set to `release` in production (Docker sets this explicitly) |

`MONGO_URI` and `JWT_SECRET` are the only two hard requirements — `config.Load()` panics on startup
if either is missing (`internal/config/config.go`).

**Generate secrets:**
```bash
openssl rand -hex 32   # JWT_SECRET, CRON_SECRET
```

Rotating `JWT_SECRET` invalidates every outstanding token (pending and access) immediately —
everyone has to log in again.

---

## `apps/web` — Next.js

| Variable | Required | Description |
|---|---|---|
| `NEXT_PUBLIC_API_URL` | Yes | Base URL of the Go server. Local: `http://localhost:8080`. Production: your Railway/Fly.io URL |

That's the entire env surface for the frontend today — there is no database URL, no auth secret, no
NextAuth config, because `apps/web` never talks to MongoDB or signs anything. It just calls the API.

---

## `infra/.env` — Docker Compose

Used only by `infra/docker-compose.yml` (local MongoDB + server + one-shot seed container).

| Variable | Required | Default | Description |
|---|---|---|---|
| `MONGO_ROOT_USER` | No | `reliva` | Mongo container root username |
| `MONGO_ROOT_PASSWORD` | Yes | — | Mongo container root password |
| `DB_NAME` | No | `reliva` | Database name (shared with the server container) |
| `JWT_SECRET` | Yes | — | Passed through to the server container |
| `SERVER_PORT` | No | `8080` | Passed through to the server container |
| `GIN_MODE` | No | `release` | Passed through to the server container |
| `ALLOWED_ORIGINS` | No | `http://localhost:3000,https://reliva.vercel.app` | Passed through |
| `CRON_SECRET` | Yes | — | Passed through to the server container |
| `VAPID_PUBLIC_KEY` / `VAPID_PRIVATE_KEY` / `VAPID_EMAIL` | No | `""` | Passed through — Phase 5 |
| `SEED_EMAIL` | Yes, to seed | — | Email for the account `docker compose run --rm seed` creates |
| `SEED_PASSWORD` | Yes, to seed | — | Password for that account |

The compose file builds the Go server from `apps/server/Dockerfile` (multi-stage: `server` target
for the API, `seed` target for the one-shot seed binary) and starts a `mongo:8` container with a
health check gating server startup.

Re-running `docker compose run --rm seed` with a different `SEED_EMAIL` creates another account
without touching existing ones — the seed command only skips if that specific email already exists.

---

## Local Development Setup

```bash
# Backend + MongoDB via Docker (recommended)
cd infra
cp .env.example .env   # fill in MONGO_ROOT_PASSWORD, JWT_SECRET, CRON_SECRET
docker compose up --build

# Frontend
bun install             # from monorepo root
cp apps/web/.env.example apps/web/.env.local
bun dev                 # http://localhost:3000
```

Or run the Go server natively instead of Docker:
```bash
cd apps/server
cp .env.example .env
go mod tidy
make run     # or: make dev (hot reload via air)
```

---

## Production Setup

- **`apps/web`** deploys to Vercel. Set `NEXT_PUBLIC_API_URL` in the Vercel dashboard
  (Production + Preview), pointing at the deployed Go server. Root Directory: `apps/web`
- **`apps/server`** deploys as a Docker container to Railway or Fly.io (needs a persistent process —
  Vercel doesn't run Go). Set all required env vars in that platform's dashboard
- **MongoDB**: point `MONGO_URI` at an Atlas cluster (or wherever Mongo is hosted) in production

---

## Security Notes

- All secrets should be unique — never reuse a value across `JWT_SECRET` / `CRON_SECRET` / Mongo
  passwords
- Rotate `JWT_SECRET` only intentionally — it logs out every user immediately
- `CRON_SECRET` and Mongo credentials should each be independently generated
- The seeded password is the account's real password, chosen at seed time — there's no forced reset
  flow beyond `POST /api/v1/auth/forgot-password` (which itself requires a backup code, so it's
  useless until TOTP has been set up once)
