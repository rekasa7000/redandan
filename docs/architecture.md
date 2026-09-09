# Architecture

## Core Principle: Standalone Go Server

The API is a separate Go binary (`apps/server`) — the single source of truth for data and business
logic. It has no knowledge of which client is consuming it.

This means:
- All data operations go through `/api/v1/**` REST endpoints served by Gin
- The server never returns HTML — pure JSON only
- `apps/web` (Next.js) is a **pure frontend**: no API routes, no MongoDB access, no server-side
  business logic. It calls the Go API exactly like any other client would
- Adding a new client (browser extension, mobile app, CLI) requires zero changes to the server

---

## System Diagram

```
                    ┌──────────────────────────────┐
                    │   MongoDB (Docker / Atlas)    │
                    │  users · contexts · tasks     │
                    │  events · credentials         │
                    │  notifications · push_subs    │
                    └──────────────┬───────────────┘
                                   │ mongo-driver/v2 (native, no ORM)
                    ┌──────────────▼───────────────┐
                    │      Reliva API (Go + Gin)    │
                    │      apps/server               │
                    │      Docker / Railway / Fly.io │
                    │                                 │
                    │  Auth: JWT bearer (all clients) │
                    │  2FA: TOTP required always      │
                    │  Owns ALL business logic + DB   │
                    └──────┬────────────┬────────────┘
                           │            │
           ┌───────────────▼──┐   ┌─────▼────────────────────────┐
           │  Web Frontend    │   │  Bearer Token Clients         │
           │  (Next.js,       │   │                               │
           │   Vercel)        │   │  ┌─────────────────────────┐  │
           │                  │   │  │  Browser Extension      │  │
           │  Pure frontend — │   │  │  (Phase 6)               │  │
           │  no DB, no API   │   │  │  chrome.storage.local    │  │
           │  routes. Stores  │   │  └─────────────────────────┘  │
           │  JWT in a cookie │   │                               │
           │  for its own use,│   │  ┌─────────────────────────┐  │
           │  sends it as     │   │  │  Mobile (Phase 7)        │  │
           │  Authorization:  │   │  │  Capacitor/Android       │  │
           │  Bearer <token>  │   │  │  Secure token storage    │  │
           └──────────────────┘   │  └─────────────────────────┘  │
                                  └───────────────────────────────┘
```

---

## Layers

### 1. Database Layer — MongoDB
- Local dev: MongoDB 8 via Docker Compose (`infra/docker-compose.yml`)
- Production: MongoDB Atlas (or any reachable MongoDB instance) via `MONGO_URI`
- Connected via the official `go.mongodb.org/mongo-driver/v2` — no ORM
- Singleton connection: `internal/db/db.go` connects once (`sync.Once`) and is reused by every handler
- Every query is written directly against the driver's typed BSON API in `internal/handlers/*.go`

### 2. API Layer — Go + Gin (`apps/server`)
- Entry point: `cmd/main.go` — loads config, connects Mongo, registers routes, runs with graceful
  shutdown on `SIGINT`/`SIGTERM`
- Routes: `internal/routes/routes.go` is the single source of truth for every endpoint and which
  middleware protects it
- Handlers: one file per resource in `internal/handlers/` (`auth.go`, `tasks.go`, `contexts.go`,
  `events.go`, `credentials.go`, `notifications.go`, `cron.go`), all sharing a `Handler` struct that
  holds the Mongo database handle and config (`internal/handlers/handler.go`)
- Returns pure JSON. No HTML, no redirects
- CORS is enforced per-request against `ALLOWED_ORIGINS` (`internal/middleware/cors.go`) — this layer
  is client-agnostic and does not know or care which frontend is calling it

### 3. Auth Layer — Two-Factor: Password + TOTP, JWT Bearer Tokens
- **Step 1:** `POST /api/v1/auth/login` — email + password (bcrypt comparison)
- **Step 2:** `POST /api/v1/auth/totp/validate` — 6-digit TOTP code (or a backup code), issued by
  `github.com/pquerna/otp`
- Every client — web, extension, mobile — authenticates the same way: a signed JWT
  (`github.com/golang-jwt/jwt/v5`) sent as `Authorization: Bearer <token>`. There is no session store
  and no cookie-based auth on the server side
- Two token types, distinguished by a `type` claim and enforced by middleware
  (`internal/middleware/auth.go`):
  - `pending` — issued after step 1 if TOTP is enabled; 5-minute expiry; only valid on
    `/auth/totp/validate` and `/auth/backup-code`
  - `access` — issued after step 2 (or immediately after step 1 if TOTP isn't set up yet); 30-day
    expiry; required on every other protected route
- The TOTP secret and backup code hashes are stored in the `users` collection (see `docs/database.md`)

### 4. Web Frontend Layer — Next.js (Pure Client)
- `apps/web` has no API routes and no MongoDB driver — it is one client among several
- Every page that touches data is a Client Component (`"use client"`) calling the Go API directly
  with `fetch`
- The access JWT returned by the Go server is stored in a plain (non-HttpOnly) cookie,
  `reliva_token`, via `lib/session.ts` — set with `document.cookie` from client-side code after login
- `middleware.ts` runs at the Next.js edge and only checks **whether** the `reliva_token` cookie is
  present, to redirect unauthenticated requests to `/login` before any page renders. It does **not**
  validate the JWT — that happens on the Go server on every actual data request
- All authenticated `fetch` calls attach the token manually as `Authorization: Bearer <token>`; the
  cookie is a Next.js-side convenience for the edge check, not an automatic transport mechanism to
  the Go server (different origin/port in development: `:3000` vs `:8080`)

### 5. Extension Layer (Phase 6 — not started)
- Will authenticate with the API using a bearer token stored in `chrome.storage.local`
- Will perform the same two-step login (password → TOTP) via its own popup UI
- Content script will detect login forms and autofill from decrypted vault credentials

### 6. Mobile Layer (Phase 7 — not started)
- Capacitor will wrap a static Next.js export
- Will authenticate using a bearer token stored in Capacitor's secure storage
- Calls the same Go API endpoints as every other client

---

## Authentication Flow (All Clients)

```
Client submits: email + password
  → POST /api/v1/auth/login
  → Server: bcrypt.CompareHashAndPassword
  → If invalid: 401 { "error": "invalid credentials" }
  → If valid and TOTP not yet enabled: issue a 30-day access token immediately
      (client should be routed to TOTP setup)
  → If valid and TOTP enabled: issue a 5-minute pending token, require step 2

Client submits: 6-digit TOTP code (or backup code)
  → POST /api/v1/auth/totp/validate   (Authorization: Bearer <pending token>)
    or POST /api/v1/auth/backup-code
  → Server: totp.Validate(code, secret) or bcrypt-compare against a stored backup code hash
  → If invalid: 401
  → If valid: issue a 30-day access token

All subsequent requests:
  → Authorization: Bearer <access token>
  → Every user-scoped query filters by the JWT's `sub` claim (the user's ObjectID) —
    this is what makes the data model multi-user-safe (see ADR-013 in technical-decisions.md)
```

---

## Data Flow Examples

### Task Creation (any client)
```
Client submits task form
  → POST /api/v1/tasks (Authorization: Bearer <token>)
  → Go handler: middleware.RequireAuth validates the JWT, sets userID in context
  → Handler binds + validates the JSON body (Gin's binding tags)
  → Inserts into MongoDB tasks collection, tagged with user_id from the JWT
  → Returns 201 with the created task document
```
*The handler logic is identical for web, extension, and mobile — only the caller differs.*

### Password Autofill (Extension, Phase 6)
```
User visits a login page
  → Content script detects <input type="password">
  → GET /api/v1/credentials (Authorization: Bearer <token>)
  → Server returns encrypted ciphertext (never plaintext)
  → Popup decrypts client-side using the master key (Web Crypto API)
  → Injects plaintext into form inputs
```

### Scheduled Notification
```
An external scheduler (Railway cron, GitHub Actions, cron container, etc.) fires
  → POST /api/v1/cron/notify (Authorization: Bearer <CRON_SECRET>)
  → middleware.RequireCron checks the header against CRON_SECRET (not a user JWT)
  → Finds tasks due today across all users
  → Inserts a notification document per task, tagged with that task's user_id
```
There is no Vercel Cron involved — the Go server isn't hosted on Vercel. Only `apps/web` deploys
there. The cron trigger is whatever scheduler sits in front of the deployed Go server.

---

## CORS Policy

| Client | Origin | Auth |
|---|---|---|
| Web frontend | `http://localhost:3000` / `https://reliva.vercel.app` | Bearer token |
| Browser extension | `chrome-extension://<id>` | Bearer token |
| Mobile (Capacitor) | Capacitor WebView / native | Bearer token |

CORS is handled entirely by the Go server (`internal/middleware/cors.go`), not the frontend:
```
Access-Control-Allow-Origin: <origin, if present in ALLOWED_ORIGINS>
Access-Control-Allow-Credentials: true
Access-Control-Allow-Methods: GET, POST, PATCH, DELETE, OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization
Access-Control-Max-Age: 86400
```
`OPTIONS` preflight requests are answered with `204` and no body. The allowed origins list comes
from the `ALLOWED_ORIGINS` env var (comma-separated). The extension's origin
(`chrome-extension://...`) is added once the extension has a stable ID after being built.

---

## Key Design Constraints

| Constraint | Rationale |
|---|---|
| Standalone Go API | Any client can be built without touching the server; `apps/web` has zero special privilege |
| TOTP on all clients | Maximum security — a password alone is not enough |
| Bearer JWT everywhere, no server sessions | One auth code path for web, extension, and mobile — nothing web-only to keep in sync |
| No ORM | Direct MongoDB queries via `mongo-driver/v2`, full control |
| Client-side crypto | Vault passwords never decryptable by the server |
| Multi-user by design | Every document is scoped by `user_id` from the JWT; the seed script can create more than one account (see ADR-013) — the operator just doesn't expose self-serve signup |
| One repo | `apps/server` (Go), `apps/web` (Next.js), `apps/extension` (Phase 6) share one monorepo |

---

## Deployment Architecture

```
apps/web    → GitHub push to main → Vercel auto-deploys (Root Directory: apps/web)
apps/server → Docker image (Dockerfile, multi-stage) → Railway or Fly.io
              Local dev: infra/docker-compose.yml (MongoDB + server + one-shot seed service)
MongoDB     → Atlas in production; Docker container (mongo:8) locally
```

The extension (Phase 6) will be built separately and loaded unpacked in the browser.
The mobile APK (Phase 7) will be built locally with Android Studio + Capacitor and sideloaded.
