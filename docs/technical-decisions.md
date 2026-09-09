# Technical Decisions

Architecture Decision Records (ADRs) for Reliva.
Each decision documents the context, the choice made, and why.

---

## ADR-001: No ORM — Use MongoDB Native Driver

**Status:** Decided

**Context:**
The project uses MongoDB. Common choices would be Mongoose (ODM) or Prisma (with MongoDB connector). Both add an abstraction layer on top of the native driver.

**Decision:**
Use the official `mongodb` npm package directly. Write all queries in TypeScript against the native driver API.

**Reasoning:**
- This is a personal project — simplicity matters more than convention
- Mongoose's schema system duplicates what TypeScript interfaces already provide
- Prisma's MongoDB connector is not as mature as its SQL connectors
- Direct queries give full access to MongoDB features (aggregation pipelines, partial indexes, etc.) without ORM limitations
- Less magic: what you write is exactly what gets sent to MongoDB

**Consequences:**
- Document shapes are defined as TypeScript interfaces in `lib/types.ts`
- Validation happens via Zod before data reaches the database
- Queries are more verbose but fully transparent

---

## ADR-002: Client-Side Encryption for Password Vault

**Status:** Decided

**Context:**
The password vault stores credentials. These could be encrypted server-side (the server holds the key) or client-side (the browser holds the key).

**Decision:**
Encrypt and decrypt credentials entirely on the client using the Web Crypto API. The server stores only ciphertext, IV, and salt. The master password never leaves the browser.

**Reasoning:**
- Server-side encryption means the server can decrypt your passwords — this is a single point of compromise
- If the MongoDB database is leaked, attacker gets only useless ciphertext
- If Vercel is compromised, same — no keys, no plaintext
- Web Crypto API is available natively in all modern browsers and Capacitor WebViews
- No dependency on a third-party crypto library

**Consequences:**
- If you forget the master password, vault data is permanently unrecoverable (by design)
- Master key is kept in memory only (sessionStorage key handle) and must be re-entered after browser restart
- The browser extension must also implement the same decryption logic
- Server cannot search or filter by credential content (only by unencrypted metadata like `site` and `username`)

---

## ADR-003: Auth.js for Authentication

**Status:** Superseded by ADR-011 — Auth.js was never shipped. The two-step password + TOTP
reasoning below still holds; the implementation is a custom Go JWT issuer instead.

**Context:**
The app needs authentication. Options: Auth.js (NextAuth v5), custom JWT implementation, Clerk, Supabase Auth.

**Decision:**
Use Auth.js (NextAuth v5) with a credentials provider for the web client.
A separate custom JWT endpoint (`POST /api/auth/token`) handles non-cookie clients.

**Reasoning:**
- Clerk and Supabase Auth are third-party services — unnecessary dependency for a personal app
- A fully custom JWT implementation is error-prone (token rotation, CSRF, secure cookies)
- Auth.js integrates natively with Next.js App Router and middleware
- The credentials provider supports username/password + TOTP as a two-step flow
- JWT sessions (not database sessions) work well with Vercel's stateless serverless functions
- The extension and mobile cannot use cookies reliably, so they get a separate bearer token endpoint that reuses the same validation logic

**Consequences:**
- User is seeded into the database at setup — no registration flow
- Session is available server-side via `auth()` and client-side via `useSession()`
- API routes must handle both cookie sessions and bearer tokens (via a shared `validateCaller()` helper)

---

## ADR-004: Single Repository for Web and Extension

**Status:** Decided

**Context:**
The browser extension is a separate artifact from the Next.js web app. It could live in a separate repository.

**Decision:**
Keep the extension in `extension/` within the main repo. It is a separate build but shares the same git history and types.

**Reasoning:**
- Shared TypeScript types between the web app and extension (e.g., credential shapes, API response types)
- One place to see all platform changes together
- Simpler for a solo developer — no need to coordinate between two repos
- The extension is relatively small and closely coupled to the API

**Consequences:**
- The extension has its own `build` script and `manifest.json`
- It does not use Next.js — it's built with a simple bundler (e.g., `esbuild`)
- Both builds are triggered from the root `package.json` if needed

---

## ADR-005: Capacitor for Mobile

**Status:** Decided (Phase 7)

**Context:**
The app needs to run on mobile. Options: React Native (rewrite), Expo, Capacitor, or PWA-only.

**Decision:**
Use Capacitor to wrap the existing Next.js web app into a native Android/iOS shell.

**Reasoning:**
- React Native requires rewriting all UI — high cost for no additional benefit for a personal app
- Capacitor wraps the existing web UI, so all design work is shared
- Capacitor provides native plugins for push notifications, haptics, and status bar — everything needed
- The app is already designed to be responsive (Tailwind mobile-first)
- PWA-only is not viable for personal Android install without sideloading complexity

**Consequences:**
- Next.js must produce a static export (`output: 'export'` in `next.config.ts`) for Capacitor to consume
- API calls go to the Vercel URL (not local) — internet connection required
- Android build requires Android Studio and Java SDK on the dev machine
- iOS build requires macOS and Xcode (not a current priority)

---

## ADR-006: Vercel Cron Jobs for Notifications

**Status:** Superseded by ADR-011 — `apps/server` is a standalone Go process, not hosted on Vercel,
so Vercel Cron cannot reach it. See ADR-011's consequences for the current mechanism.

**Context:**
Scheduled notifications (daily digest, deadline reminders) need a mechanism to trigger on a schedule. Options: Vercel Cron Jobs, a separate cron server, database-triggered functions, or a third-party scheduler.

**Decision:**
Use Vercel Cron Jobs configured in `vercel.json`.

**Reasoning:**
- No additional infrastructure required — Vercel already hosts the app
- Cron Jobs call a protected API route (`/api/cron/notify`) with a secret header
- Free tier includes up to 2 cron jobs running daily
- Simple, auditable, and co-located with the app

**Consequences:**
- Cron is limited to daily granularity on Vercel's free tier
- The cron route must be secured with a `CRON_SECRET` header check to prevent unauthorized triggering
- If Vercel is down at the scheduled time, the notification is missed (acceptable for personal use)

---

## ADR-007: Manifest V3 for Browser Extension

**Status:** Decided

**Context:**
Browser extensions can use Manifest V2 (legacy, being deprecated) or Manifest V3 (current standard).

**Decision:**
Build the extension targeting Manifest V3.

**Reasoning:**
- Manifest V2 is deprecated in Chrome (phased out) and will be removed
- Firefox now supports Manifest V3 (with some differences, but manageable)
- MV3 is the future-proof choice
- The extension's needs (content script autofill, background service worker, storage) are all well-supported in MV3

**Consequences:**
- Background pages become service workers in MV3 — they can be terminated by the browser
- Persistent state must use `chrome.storage` rather than in-memory variables
- The content script security model is stricter — CSP must be configured correctly
- Firefox MV3 compatibility requires testing (minor API differences)

---

## ADR-008: Route Groups for Auth-Protected Pages

**Status:** Decided

**Context:**
Some pages require authentication (all app pages), and some do not (login). Next.js App Router needs a clean way to apply the auth guard.

**Decision:**
Use route groups: `(auth)` for public auth pages, `(app)` for all protected pages. Middleware applies the auth check to all routes except those under `(auth)`.

**Reasoning:**
- Route groups (parentheses folders) are invisible in the URL
- Clean separation between authenticated and unauthenticated surfaces
- Middleware runs at the edge, before any rendering — minimal latency for auth checks
- No need for per-page auth checks — the middleware handles it globally

**Consequences:**
- All pages must be placed under the correct route group
- Middleware must maintain an updated list of public routes (or use a path-based pattern)

---

## ADR-009: Standalone API Design — Server Independent of Clients

**Status:** Decided

**Context:**
The project has three client surfaces (web, extension, mobile). The initial design treated the Next.js
API routes as tightly coupled to the web frontend. The question: should the API be designed as a
proper standalone backend that any client can consume independently?

**Decision:**
The API is the standalone backend. All three clients consume it identically via REST.
The web frontend is not privileged — it does not have special server-side data access that mobile or
extension cannot replicate. Any new client can be built by implementing the auth flow and calling the
same REST endpoints.

*Implementation note (superseding the original text below): this was first planned as Next.js Route
Handlers under `/app/api/**` sharing the Next.js process. It shipped instead as a fully separate Go
binary (`apps/server`) — see ADR-011. The principle here (client-agnostic, self-contained API) is
unchanged; only the runtime changed.*

**Reasoning:**
- A single user personal app with three frontends naturally benefits from a clean server/client split
- If a fourth client is ever needed (CLI, desktop app, another mobile platform), zero server changes are required
- Forces better API design — endpoints must be self-contained, documented, and not assume client context
- Separates concerns clearly: the server owns data and business logic, clients own presentation
- Aligns with the long-term potential of the project becoming a multi-client platform

**Consequences:**
- `apps/web` never calls MongoDB directly, not even as an RSC optimization — it has zero DB access,
  full stop (stricter than originally planned)
- Every client, including the web frontend, authenticates with a bearer JWT — there is no
  cookie-based session on the server side (see ADR-011)
- CORS is enforced entirely by the Go server (`internal/middleware/cors.go`)
- API documentation (`docs/api.md`) is the authoritative contract; `openapi.yaml` is the
  machine-readable version of it

---

## ADR-010: TOTP (Authenticator App) as Mandatory Second Factor

**Status:** Decided

**Context:**
The app holds sensitive personal data — tasks, calendar events, and an encrypted password vault.
A password alone is a single point of failure: if the password is guessed, phished, or leaked,
everything is exposed. The question: require a second factor, and if so, which kind?

Options: SMS OTP, email OTP, TOTP (authenticator app), hardware key (WebAuthn/FIDO2), passkeys.

**Decision:**
Require TOTP (Time-based One-Time Password) as a mandatory second factor on every login, for all
clients. Implemented server-side with `otplib`. Setup via a QR code scanned into Google Authenticator,
Authy, or any TOTP-compatible app.

**Reasoning:**
- SMS OTP requires a phone number and a third-party SMS provider — unnecessary infrastructure
- Email OTP requires access to email during login — circular dependency if email credentials are in the vault
- TOTP is offline — the authenticator app generates codes without internet or servers
- TOTP is widely supported — any TOTP app (Google Authenticator, Authy, Bitwarden, 1Password) works
- WebAuthn / passkeys are the most secure option but add implementation complexity;
  TOTP is the pragmatic high-security choice for a personal app
- `otplib` is a well-maintained, standards-compliant library; no need to implement RFC 6238 manually

**Consequences:**
- Login is a two-step flow: (1) password → (2) TOTP code
- The TOTP secret is generated once at setup and stored encrypted in the `users` collection
- Backup codes are generated at setup (10 single-use codes) and must be stored offline by the user
- The login page must handle the two-step UI gracefully
- The extension and mobile login popups must also implement the two-step flow
- If the authenticator app and backup codes are both lost, account access requires manual DB intervention

*Implementation note: shipped as 8 backup codes (`generateBackupCodes(8)` in
`internal/handlers/auth.go`), not 10. The TOTP secret is stored in plaintext in the `users`
collection, not encrypted — see the note in ADR-011's consequences.*

---

## ADR-011: Go + Gin Standalone Server Replaces Next.js API Routes and Auth.js

**Status:** Decided (implemented in Phase 1)

**Context:**
ADR-003 and ADR-009 originally planned the API as Next.js Route Handlers under `/app/api/**`,
authenticated via Auth.js (NextAuth v5) for the web client and a separate custom JWT endpoint for
extension/mobile. While building Phase 1, this was reconsidered before any Auth.js code was written.

**Decision:**
Build the API as a fully separate Go + Gin binary (`apps/server`), with its own auth system — a
custom JWT issuer (`golang-jwt/jwt/v5`) — used identically by every client, including the web
frontend. `apps/web` becomes a pure frontend: no API routes, no MongoDB driver, no server-side auth
framework. Every page that needs data is a Client Component calling the Go API with `fetch`.

**Reasoning:**
- A literal separate process enforces "standalone API" (ADR-009) far more strongly than Route
  Handlers ever could — there is no way for `apps/web` to accidentally gain a shortcut into the
  database, because it has no MongoDB driver installed at all
- One auth code path instead of two: Auth.js for cookies + a bonus JWT endpoint for everyone else
  meant maintaining two authentication systems that had to independently enforce the same TOTP
  requirement. A single JWT issuer, used by every client, removes that duplication entirely
- Go's standard library + Gin is a smaller, more auditable surface for something as
  security-sensitive as auth and a zero-knowledge vault than a NextAuth credentials-provider
  integration
- Deploying the API independently of the frontend (Railway/Fly.io vs Vercel) means the frontend can
  be redeployed without ever touching the database layer, and vice versa

**Consequences:**
- Two independent deploy pipelines instead of one (see `docs/environment.md`)
- The web frontend stores its JWT in a plain (non-HttpOnly) cookie, `reliva_token`, purely so
  Next.js edge middleware can check token *presence* before rendering — this is weaker than an
  HttpOnly session cookie would be against XSS, but is mitigated by the token being short-lived
  server-side-invalidatable only via `JWT_SECRET` rotation. This is a known, accepted trade-off for
  a personal single-operator app; revisit before ever exposing this to untrusted users (see ADR-012)
- The TOTP secret is currently stored in plaintext in the `users` collection (`totp_secret` field),
  not encrypted as ADR-010 originally specified — worth fixing before this app ever holds
  higher-stakes data
- `POST /api/v1/cron/notify` (formerly `/api/cron/notify`) is triggered by whatever external
  scheduler sits in front of the deployed Go server — Railway cron, a scheduled GitHub Action, or
  similar. There is no Vercel Cron config for it (supersedes ADR-006)
- `docs/architecture.md`, `docs/api.md`, `docs/database.md`, `docs/environment.md`, and
  `docs/tech-stack.md` describe the Go implementation directly rather than the original Next.js plan

---

## ADR-012: Multi-User Data Model, Single Operator by Choice

**Status:** Decided

**Context:**
Phase 1 shipped with the seed script (`cmd/seed/main.go`) refusing to run if **any** user already
existed in the `users` collection — a hard single-account gate. Every actual data query, however, was
already written to filter by `user_id` extracted from the JWT (`internal/handlers/*.go`), because
that's simply how per-caller data isolation works with a shared JWT-authenticated API. The operator
asked to add a second, real account after the placeholder seed account — surfacing that the seed
script's gate was the only thing standing in the way.

**Decision:**
Keep the data model and every handler exactly as they were — they were already multi-user-safe.
Change the seed script's existence check from "does any user exist" to "does this specific email
already exist," so it can be re-run to create additional accounts. Do not add a self-serve signup
endpoint, an invite system, or per-account admin roles — those are unneeded complexity for what is
still, in practice, an app one person (the operator) uses.

**Reasoning:**
- The `user_id`-scoped query pattern was already the correct, general-purpose way to isolate data per
  caller — building it as if there were exactly one user (e.g. a global "the user" singleton) would
  have been the unnecessary special case, not the other way around
- The only real constraint was administrative (the seed script's blanket skip), not architectural —
  fixing it is a small, low-risk change
- A signup endpoint, invite flow, or per-account roles/permissions would be real added complexity
  with no current use case — `docs/overview.md`'s "engineered for one person" framing is still true
  in practice, it's just no longer a hard technical ceiling

**Consequences:**
- `docker compose run --rm seed` (with `SEED_EMAIL`/`SEED_PASSWORD` set in `infra/.env`) can be run
  once per account to create
- Each account gets its own default contexts (Personal/Work/Health) and fully isolated tasks,
  events, credentials, and notifications — nothing is shared across accounts
- There is still no way for a user to create their own account through the UI — every account is
  provisioned by whoever has shell/Docker access to the deployment. This is intentional, not a gap
- `users.email` has no unique index in code (see `docs/database.md`) — uniqueness is only enforced
  at seed time. Acceptable for the current provisioning model; would need a proper unique index if
  a signup endpoint is ever added
- PLANNING.md's tagline changed from "Single-user. Personal." to "Multi-user by design. Personal in
  practice."
