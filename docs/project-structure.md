# Project Structure

## Monorepo Layout

Reliva uses a **bun workspaces** monorepo for the JS packages, with a separate Go module for the
server. The root holds workspace configuration, shared tooling, and documentation.

```
reliva/                         ← Monorepo root
  apps/
    web/                          ← Next.js pure frontend (@reliva/web)
    server/                       ← Go + Gin standalone API (module: reliva/server)
    extension/                    ← Browser extension (@reliva/extension) — Phase 6, not started
  infra/                          ← docker-compose for local MongoDB + server + seed
  docs/                           ← Documentation (this folder)
  scripts/                        ← seed.ts / seed.cjs — legacy, superseded by apps/server/cmd/seed
  package.json                    ← Workspace root — scripts delegate to apps/web
  bun.lock                        ← Shared lockfile across JS workspaces
  openapi.yaml                    ← API contract, source for generated TS types
  .gitignore
  .prettierrc
  .prettierignore
  AGENTS.md
  PLANNING.md
  .env.example
```

Workspaces are declared in the root `package.json`:
```json
{ "workspaces": ["apps/*"] }
```
`apps/server` is a Go module, not a bun workspace — `bun install` skips it. Go dependencies are
managed independently via `go.mod`/`go.sum`.

---

## Current State (Phase 1 complete)

```
reliva/
  apps/
    web/
      app/
        (auth)/
          login/page.tsx           ← Two-step login (password → TOTP/backup code)
          forgot-password/page.tsx
        (app)/
          layout.tsx                ← Auth guard shell (middleware.ts does the actual redirect)
          page.tsx                  ← Dashboard — placeholder, "Phase 2 coming soon"
          settings/page.tsx         ← TOTP setup (QR + backup codes), change password, logout
        favicon.ico
        globals.css
        layout.tsx
      components/
        ui/                        ← shadcn/ui primitives (button, card, input, label)
        theme-provider.tsx
      hooks/                       ← empty (.gitkeep)
      lib/
        session.ts                 ← reliva_token cookie helpers (get/set/clear)
        types.gen.ts                ← generated from openapi.yaml — do not hand-edit
        utils.ts                    ← cn() utility
      public/
      middleware.ts                 ← Edge auth guard — checks token presence, not validity
      components.json               ← shadcn config
      eslint.config.mjs
      next.config.ts
      package.json                  ← @reliva/web
      postcss.config.mjs
      tsconfig.json
      vercel.json
    server/
      cmd/
        main.go                     ← Entry point, graceful shutdown
        seed/main.go                ← One-shot account seeder (per-email, not per-install)
      internal/
        config/config.go            ← Env var loading, panics on missing required vars
        db/db.go                    ← MongoDB singleton (sync.Once)
        models/models.go            ← All bson/json document structs
        middleware/
          auth.go                   ← RequireAuth, RequirePending, RequireCron
          cors.go                   ← CORS + StripTrailingSlash
        handlers/
          handler.go                ← Base Handler struct (db + config)
          auth.go                   ← Login, TOTP setup/confirm/validate, backup codes, change/forgot password
          tasks.go                  ← Full CRUD
          contexts.go                ← Full CRUD
          events.go                  ← Full CRUD
          credentials.go             ← Zero-knowledge vault CRUD
          notifications.go           ← List, mark-read, push subscribe
          cron.go                    ← Cron notify handler
        routes/routes.go             ← Single source of truth for routes + middleware
      .air.toml                      ← Hot-reload config (air)
      .env.example
      Dockerfile                     ← Multi-stage: `server` and `seed` build targets
      go.mod / go.sum
      Makefile                       ← dev, run, build, tidy targets
    extension/
      README.md                      ← Phase 6 placeholder, no code yet
  infra/
    docker-compose.yml               ← mongo + server + one-shot seed service
    .env.example
  docs/
  scripts/
    seed.ts / seed.cjs               ← legacy Node seed script, not wired into any package.json script
  openapi.yaml
  package.json
  bun.lock
```

Not yet built (Phase 2+): task/calendar/vault pages, `lib/api.ts` fetch wrapper, `lib/validations.ts`,
sidebar/nav components, and everything under `apps/extension`.

---

## What Changed From the Original Plan

Earlier drafts of this document (and `PLANNING.md`) described `apps/web` owning `app/api/**` Next.js
Route Handlers, Auth.js session cookies, and a direct `lib/db.ts` MongoDB connection. That plan was
abandoned during Phase 1 in favor of a standalone Go server — see ADR-011 in
`docs/technical-decisions.md`. None of the following ever existed in `apps/web` and should not be
recreated there: `app/api/`, `lib/db.ts`, `lib/auth.ts` (Auth.js config), `lib/totp.ts`. TOTP, auth,
and all database access live exclusively in `apps/server`.

`apps/web/lib/crypto.ts` (PBKDF2 + AES-GCM for the vault) is still correctly planned for Phase 4 —
that part of the architecture didn't change, since vault encryption was always meant to be
client-side only.

---

## Target State (Remaining Phases)

```
apps/web/
  app/
    (auth)/
      login/page.tsx                ← done
      forgot-password/page.tsx      ← done
    (app)/
      layout.tsx                     ← done
      page.tsx                       ← Phase 2: real dashboard
      settings/page.tsx              ← done
      tasks/
        page.tsx
        new/page.tsx
        [id]/page.tsx
      calendar/page.tsx              ← Phase 3
      vault/
        page.tsx                     ← Phase 4
        new/page.tsx
        [id]/page.tsx
  components/
    ui/                              ← done
    shared/                          ← sidebar, bottom-nav — Phase 2
    tasks/                           ← Phase 2
    calendar/                        ← Phase 3
    vault/                           ← Phase 4
  lib/
    session.ts                       ← done
    types.gen.ts                     ← done (regenerate via `bun run gen:types`)
    api.ts                           ← Phase 2: typed fetch wrapper around the Go API
    validations.ts                   ← Phase 2: Zod schemas for forms
    crypto.ts                        ← Phase 4: Web Crypto (client-side only)
    utils.ts                         ← done

apps/extension/                      ← Phase 6
  manifest.json
  popup/{index.html,popup.ts}
  background/service-worker.ts
  content/autofill.ts

apps/mobile/                         ← Phase 7, new workspace when added
  android/ ios/ capacitor.config.ts
```

`apps/server` is already feature-complete for Phases 1–5's API surface (auth, tasks, contexts,
events, credentials, notifications, push subscribe, cron) — remaining phase work is almost entirely
in `apps/web` and `apps/extension`.

---

## Adding a New App (e.g. Mobile — Phase 7)

```
apps/
  web/
  server/
  extension/
  mobile/               ← new: @reliva/mobile
    android/
    ios/
    capacitor.config.ts
    package.json
```
Run `bun install` at the root to pick up the new workspace. `apps/server` needs no changes — mobile
calls the same API.

---

## Deployment Targets

- **`apps/web`** → Vercel. Project Root Directory: `apps/web`. Vercel runs `bun install` at the
  monorepo root, then `bun run build` inside `apps/web`
- **`apps/server`** → Docker image built from `apps/server/Dockerfile`, deployed to Railway or
  Fly.io. Vercel cannot run this — it needs a persistent Go process, not a serverless function
- **MongoDB** → Atlas in production, `mongo:8` in Docker locally

These are two independent deployments. There is no single "deploy" step that covers both.

---

## Key File Responsibilities

| File | Responsibility |
|---|---|
| `apps/server/cmd/main.go` | Go entry point. Loads config, connects MongoDB, registers routes, runs with graceful shutdown. |
| `apps/server/cmd/seed/main.go` | Creates one user account + default contexts. Skips only if that specific email already exists — safe to run per-account. |
| `apps/server/internal/config/config.go` | Reads all env vars; panics on missing required vars so misconfiguration fails fast. |
| `apps/server/internal/db/db.go` | MongoDB singleton — connect once, reuse across handlers. |
| `apps/server/internal/middleware/auth.go` | JWT validation. `RequireAuth` (access token), `RequirePending` (step-1 token), `RequireCron` (secret). |
| `apps/server/internal/handlers/auth.go` | Two-step login: password → pending JWT → TOTP → access JWT. Also TOTP setup and backup codes. |
| `apps/server/internal/routes/routes.go` | Single source of truth for all routes and which middleware protects each group. |
| `apps/web/lib/session.ts` | Reads/writes the `reliva_token` cookie. Not a validation mechanism — just client-side persistence. |
| `apps/web/middleware.ts` | Edge middleware — redirects based on token *presence*, not validity. |
| `apps/web/lib/types.gen.ts` | Generated from `openapi.yaml` via `bun run gen:types`. Never hand-edit. |
| `openapi.yaml` (root) | Authoritative API contract — source of truth for `types.gen.ts`. |
| `package.json` (root) | Workspace root. Scripts delegate to `apps/web`. Prettier lives here. |

---

## Naming Conventions

| Item | Convention | Example |
|---|---|---|
| Files (web) | kebab-case | `task-card.tsx` |
| React components | PascalCase | `TaskCard` |
| TS functions / variables | camelCase | `getTaskById` |
| Go files | snake_case-free, package-grouped | `internal/handlers/tasks.go` |
| Go functions | PascalCase (exported) / camelCase (unexported) | `ListTasks`, `parseUserAndResourceID` |
| Package names (JS) | `@reliva/<name>` | `@reliva/web` |
| API routes | plural noun, versioned | `/api/v1/tasks`, `/api/v1/contexts` |
| MongoDB collections | plural noun, snake_case | `tasks`, `push_subscriptions` |
| Environment variables | SCREAMING_SNAKE_CASE | `MONGO_URI` |

---

## Running the Project

```bash
# Backend + MongoDB (Docker, recommended)
cd infra
cp .env.example .env
docker compose up --build

# Frontend (from monorepo root)
bun install
bun dev                          # delegates to apps/web

# Backend without Docker
cd apps/server
cp .env.example .env
go mod tidy
make run                         # or: make dev (hot reload via air)
```
