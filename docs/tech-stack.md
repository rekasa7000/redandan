# Tech Stack

## Summary Table

| Layer | Technology | Version | Role |
|---|---|---|---|
| **Backend** ||||
| Language | Go | 1.25 | API server — all business logic, all DB access |
| Framework | Gin | v1.12 | HTTP routing and middleware |
| Database | MongoDB | 8.x | Primary data store |
| DB Driver | `go.mongodb.org/mongo-driver/v2` | v2.5 | Direct queries, no ORM |
| Auth tokens | `golang-jwt/jwt/v5` | v5.2 | Signs pending/access bearer JWTs |
| 2FA | `pquerna/otp` | v1.4 | TOTP generation and validation (RFC 6238) |
| QR Code | `skip2/go-qrcode` | — | QR code PNG generation for TOTP setup |
| Password hashing | `golang.org/x/crypto/bcrypt` | v0.35 | Password + backup code hashing |
| Env loading | `joho/godotenv` | v1.5 | Loads `.env` in local dev only |
| Deployment | Docker → Railway or Fly.io | — | Persistent process (Vercel doesn't run Go) |
| **Frontend** ||||
| Framework | Next.js | 16.x | Pure frontend — no API routes, no DB access |
| Language | TypeScript | 5.x | Type safety |
| Runtime | React | 19.x | UI |
| UI Library | shadcn/ui | latest | Component library on top of Radix |
| Styling | Tailwind CSS | 4.x | Utility-first CSS |
| Type generation | `openapi-typescript` | 7.x | Generates `lib/types.gen.ts` from `openapi.yaml` |
| Package Manager | bun | latest | Installs, scripts, dev server |
| Deployment | Vercel | — | Hosting only — no serverless functions used |
| **Shared / Future** ||||
| Validation | Zod | — | Planned for form + client-side validation (Phase 2+) |
| Push | Web Push (VAPID) | — | Phase 5 |
| Mobile | Capacitor | — | Phase 7 |
| Extension | Manifest V3 | — | Phase 6 |

---

## Backend — Go + Gin (`apps/server`)

The entire API, auth system, and database access live in one Go binary. There is no framework
splitting "API routes" from "the app" — Gin registers every route in
`internal/routes/routes.go`, and each handler in `internal/handlers/` does validation, business
logic, and the Mongo query in one function.

Why Go instead of keeping everything in Next.js:
- A genuinely standalone API that isn't tied to a JS runtime or a serverless function's request
  lifecycle
- Runs as a persistent process — no cold starts, straightforward to reason about for a small,
  latency-insensitive personal app
- Deployable independently of the frontend (Railway/Fly.io vs Vercel)

---

## Database — MongoDB (native Go driver)

Same rationale as before the migration: tasks, events, and credentials have different shapes and
optional fields, which maps naturally to documents. `go.mongodb.org/mongo-driver/v2` is used
directly — no Mongoose-equivalent, no code generation, no schema migration tool. Document shapes are
plain Go structs with `bson`/`json` tags in `internal/models/models.go`.

Local dev uses a Dockerized `mongo:8` container (`infra/docker-compose.yml`). Production points
`MONGO_URI` at Atlas or any reachable Mongo instance.

---

## Auth — Custom JWT + TOTP (no Auth.js)

There is no session framework. Every client — including the web frontend — authenticates the same
way:

1. `POST /api/v1/auth/login` (email + password) → pending or access JWT
2. `POST /api/v1/auth/totp/validate` (TOTP code) → access JWT
3. Every subsequent request: `Authorization: Bearer <access JWT>`

Tokens are signed HS256 JWTs (`golang-jwt/jwt/v5`) with a `type` claim (`pending` | `access`) and
standard registered claims (`sub`, `iat`, `exp`). `pquerna/otp` implements TOTP (RFC 6238) for 2FA;
`skip2/go-qrcode` renders the setup QR code server-side as a PNG data URL so the frontend needs no QR
library.

The web frontend persists its JWT in a plain (non-HttpOnly) `document.cookie` purely so Next.js edge
middleware can check token *presence* before rendering a protected page — the middleware never
validates the JWT itself. Real validation happens on the Go server on every data request. See
`docs/architecture.md` for the full flow.

---

## UI — shadcn/ui + Tailwind CSS v4

Unchanged by the backend migration. shadcn/ui components are copied into `apps/web/components/ui/`
and owned directly — no npm package to version. Tailwind v4 uses CSS-first configuration (no
`tailwind.config.js`).

**Design constraint: no gradients.** Flat, solid colors from the theme only.

---

## Type Safety — OpenAPI → TypeScript

`openapi.yaml` at the monorepo root is the authoritative API contract. Running
```bash
bun run gen:types
```
regenerates `apps/web/lib/types.gen.ts` via `openapi-typescript`. This is how the frontend gets typed
request/response shapes without hand-writing interfaces that can drift from the Go structs.

---

## Cryptography — Web Crypto API (Vault, Phase 4)

Unchanged: the password vault will use the browser-native Web Crypto API entirely client-side.
1. User enters master password
2. A `CryptoKey` is derived using PBKDF2 with a random salt
3. Each credential is encrypted with AES-GCM (256-bit key, random IV per item)
4. Only ciphertext, IV, and salt are sent to the Go server and stored in MongoDB
5. Decryption happens entirely in the browser — the server never sees plaintext

---

## Push Notifications — web-push (VAPID, Phase 5)

Not yet implemented. When it lands: VAPID keys identify the server to push services, subscriptions
are stored in `push_subscriptions` (already modeled and has a working `POST /api/v1/push/subscribe`
endpoint), and an external scheduler calls `POST /api/v1/cron/notify` on the Go server — there is no
Vercel Cron involved since the API isn't hosted on Vercel.

---

## Package Manager — bun (frontend only)

`apps/web` and the monorepo root use bun (`bun.lock`). `apps/server` is Go — it uses `go.mod`/`go.sum`
and the standard `go` toolchain (`go mod tidy`, `go build`), not bun.

---

## Deployment

| App | Platform | Trigger |
|---|---|---|
| `apps/web` | Vercel | Push to `main`, Root Directory `apps/web` |
| `apps/server` | Docker → Railway or Fly.io | Manual/CI build of `apps/server/Dockerfile` |
| MongoDB | Atlas (prod) / Docker `mongo:8` (local) | — |

There are two independent deploy pipelines, not one. A change to `apps/server` does not trigger a
Vercel deploy and vice versa.
