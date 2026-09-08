# Reliva

Personal operating system for your life. Tasks, calendar, password vault, push notifications — one private system, three clients.

> **re** (personal brand) · **li** (literal) · **va** (virtual assistant)

## What it does

- **Task tracker** — tasks organized by job/life context (work, personal, freelance, etc.)
- **Calendar & events** — payroll dates, deadlines, vacations in one timeline
- **Password vault** — zero-knowledge: AES-GCM encryption happens client-side only; the server never sees plaintext
- **Push notifications** — daily reminders for due tasks via Web Push
- **Browser extension** — autofill from the vault, quick task creation (Phase 6)
- **Mobile app** — Capacitor wrapper for Android (Phase 7)

## Monorepo layout

```
reliva/
  apps/
    web/        — Next.js 16 + Auth.js  (@reliva/web, deployed to Vercel)
    server/     — Go 1.23 + Gin         (standalone REST API, deployed to Railway/Fly.io)
    extension/  — Manifest V3           (@reliva/extension, Phase 6)
  infra/        — docker-compose for local MongoDB + server
  docs/         — architecture, ADRs, API reference, roadmap
```

## Getting started

### Prerequisites

- [Bun](https://bun.sh) — JavaScript package manager
- [Go 1.23+](https://go.dev/dl/) — for the backend
- [Docker](https://www.docker.com) — for local development with MongoDB

### Run locally with Docker (recommended)

```bash
cd infra
cp .env.example .env   # fill in MONGO_ROOT_PASSWORD, JWT_SECRET, CRON_SECRET
docker compose up --build
```

The API will be available at `http://localhost:8080`.

### Run the backend without Docker

```bash
cd apps/server
cp .env.example .env   # fill in MONGO_URI, JWT_SECRET, etc.
go mod tidy
make dev               # hot-reload via air (go install github.com/air-verse/air@latest)
# or
make run               # no hot-reload
```

### Run the web app

```bash
bun install            # from monorepo root
bun dev                # starts apps/web on http://localhost:3000
```

## Tech stack

| Layer | Technology |
|---|---|
| Web frontend | Next.js 16, React 19, TypeScript, Tailwind CSS v4, shadcn/ui |
| Backend API | Go 1.23, Gin, MongoDB driver v2 |
| Database | MongoDB |
| Auth | Auth.js (web) · JWT bearer tokens (extension/mobile) · TOTP mandatory 2FA |
| Vault encryption | Web Crypto API — PBKDF2 + AES-GCM, client-side only |
| Push notifications | Web Push (VAPID) |
| Deployment | Vercel (web) · Railway or Fly.io (server) |

## Security

Authentication is two-step and mandatory:
1. Password → short-lived pending token (5 min)
2. TOTP code (or backup code) → 30-day access token

The password vault uses zero-knowledge encryption. The server stores only ciphertext + IV + salt and has no ability to decrypt.

## Docs

Full architecture, ADRs, API reference, and roadmap live in [`docs/`](docs/README.md).
