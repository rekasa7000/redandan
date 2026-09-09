# Security

Security is a first-class concern in Reliva — especially for the password vault and authentication.
This document covers the threat model, the mitigations, and the rules that must not be broken.

---

## Threat Model

Reliva is a personal app, provisioned per-account by whoever controls the deployment (see ADR-012 in
`docs/technical-decisions.md`) — in practice, one operator. The relevant threats are:

| Threat | Likelihood | Impact |
|---|---|---|
| Unauthorized access to the web app | Medium | High — all data exposed |
| Password stolen or guessed | Low-Medium | Mitigated by TOTP (single factor not enough) |
| TOTP code intercepted or stolen | Very Low | High — valid only for 30 seconds |
| Database breach (MongoDB) | Low | Critical for vault if not encrypted |
| Bearer token stolen (any client) | Low-Medium | High — full API access for that account until expiry |
| `reliva_token` cookie read via XSS | Low | High — the cookie is not HttpOnly (see below) |
| Cron endpoint triggered by attacker | Low | Low (only creates notification records) |
| Weak master password for vault | Medium (user error) | Critical — vault decryptable |
| Authenticator app + backup codes both lost | Low (user error) | Account lockout |

---

## Authentication — Two-Factor (Password + TOTP), Custom JWT

Authentication requires **two independent factors** on every login. Passing one does not grant
access. There is no Auth.js/NextAuth involved — the Go server (`apps/server`) issues and validates
its own JWTs, and every client (web, extension, mobile) authenticates identically.

### Step 1 — Password

- `POST /api/v1/auth/login`
- The stored password is hashed with **bcrypt** (`golang.org/x/crypto/bcrypt`, default cost) in the
  `users` collection
- On login, the submitted password is compared with `bcrypt.CompareHashAndPassword`
- A failed password check returns `401 { "error": "invalid credentials" }` — no information about
  which factor failed
- **Not yet implemented:** rate limiting on failed login/TOTP attempts. There is no IP-based lockout
  today. Worth adding before this is exposed beyond a trusted network

### Step 2 — TOTP (Time-Based One-Time Password)

- `POST /api/v1/auth/totp/validate`, requires a `pending` token from step 1
- The user enters the 6-digit code from their authenticator app
- The server validates using `pquerna/otp`'s `totp.Validate(code, secret)` (RFC 6238, ±1 time step for clock drift)
- **Known gap:** there is no replay-prevention cache — a valid code could theoretically be reused
  within its 30-second window if intercepted. Low real-world risk for a personal deployment, but not
  what ADR-010 originally assumed

### Session / Token Issuance

After both factors pass, every client receives the same thing — a signed JWT, not a session:

| Client | Storage | Expiry |
|---|---|---|
| Web browser | Plain (non-HttpOnly) cookie `reliva_token`, set via `document.cookie` in `lib/session.ts` | 30 days |
| Browser extension (Phase 6) | `chrome.storage.local` | 30 days |
| Mobile app (Phase 7) | Capacitor secure storage | 30 days |

All tokens are HS256-signed with `JWT_SECRET`. The payload is a standard JWT claims set plus a
`type` field (`pending` or `access`) — readable by anyone with the token, tamper-proof but not
confidential. Do not put secrets in the JWT payload.

**The web cookie is not HttpOnly.** It has to be readable by client-side JavaScript because the
frontend attaches it manually as `Authorization: Bearer <token>` on every API call — the Go server
and the Next.js dev server are different origins/ports, so the browser won't send the cookie
automatically anyway. The cookie's only job on the Next.js side is letting `middleware.ts` check
*presence* at the edge before rendering a page. This is a deliberate trade-off (see ADR-011) that
trades some XSS resistance for a single, uniform bearer-token auth model across all clients. If this
app ever serves untrusted users, revisit this.

### TOTP Setup (One-Time, Per Account)

1. `GET /api/v1/auth/totp/setup` — generates a TOTP secret via `pquerna/otp`, stores it as
   `totp_pending_secret` on the user document, returns a QR code (`skip2/go-qrcode`, PNG data URL)
   and the raw secret for manual entry
2. The user scans the QR code or enters the secret manually into an authenticator app
3. `POST /api/v1/auth/totp/confirm` — verifies a code against the pending secret; on success, moves
   it to `totp_secret`, sets `totp_enabled: true`, and generates 8 single-use backup codes (shown
   once in the response)

**Known gap:** `totp_secret` is stored in **plaintext** in the `users` collection, not encrypted.
Earlier planning assumed encryption at rest keyed off an app secret; that was never implemented in
the Go server. Anyone with direct database access can read active TOTP secrets. Acceptable for the
current threat model (single operator, trusted DB access) but should be fixed — e.g. AES-GCM with a
key derived from `JWT_SECRET` or a dedicated secret — before broadening access to the deployment.

### Backup Codes

- 8 codes generated at TOTP confirm time (not 10 — see above), each 12 hex characters
  (`generateBackupCodes` in `internal/handlers/auth.go`)
- Stored as bcrypt hashes in `users.backup_codes` — plaintext never stored
- `POST /api/v1/auth/backup-code` (login step 2 alternative) and
  `POST /api/v1/auth/forgot-password` (password reset) both burn the code on successful use
- Once exhausted, there is no self-serve regeneration endpoint yet — recovery requires re-running
  TOTP setup (which requires being logged in) or direct database access

**Rules:**
- `JWT_SECRET` must be long and random (32+ bytes). Generate: `openssl rand -hex 32`
- Every seeded password must be strong and unique
- Never commit `apps/server/.env`, `apps/web/.env.local`, or `infra/.env`
- Store backup codes offline

---

## Password Vault — Zero-Knowledge Design (Phase 4, not yet built)

The most security-critical module. The design ensures the server has zero ability to decrypt vault
contents — this part of the plan is unchanged by the Go migration.

### Encryption Flow (Client → Server)

```
User types master password
  → PBKDF2 key derivation (100,000 iterations, SHA-256, random 16-byte salt)
  → Produces a CryptoKey (AES-GCM 256-bit)
  → Kept in memory only (never stored, never sent)

User adds a credential
  → Generate random 12-byte IV
  → AES-GCM encrypt plaintext password with derived key + IV
  → POST /api/v1/credentials with { encrypted_data, iv, salt, username, site }

Server stores verbatim
  → MongoDB receives: { encrypted_data, iv, salt }
  → Go handler never sees plaintext, never derives the key
```

### Decryption Flow (Server → Client)

```
User opens vault
  → Prompted for master password (re-derives key in memory)
  → GET /api/v1/credentials returns ciphertext + iv + salt
  → Client decrypts each entry with the derived key
  → Plaintext displayed in browser memory only
```

### Key Rules — Never Break These

1. **Never add server-side decryption logic to `apps/server`.** The Go handlers return ciphertext
   and that's it — see `internal/handlers/credentials.go`.
2. **Never log credential plaintext** anywhere — not in Go handlers, not in the browser console.
3. **Never store the master password or derived key in `localStorage`, a cookie, or MongoDB.**
   Memory only, on the client.
4. **The browser extension must implement its own decryption logic** — never send the derived key
   from the web app to the extension.
5. **If the master password is lost, the vault data is unrecoverable.** By design.

### Cryptographic Parameters (planned for `apps/web/lib/crypto.ts`, Phase 4)

| Parameter | Value |
|---|---|
| Key derivation | PBKDF2 |
| Hash | SHA-256 |
| Iterations | 100,000 |
| Salt length | 16 bytes (random per credential) |
| Encryption | AES-GCM |
| Key length | 256 bits |
| IV length | 12 bytes (random per credential) |

Do not change these without a migration plan for existing vault data.

---

## API Security

### Auth Middleware

Every protected route in `internal/routes/routes.go` is wrapped in one of:
- `middleware.RequireAuth(jwtSecret)` — valid `access` token required
- `middleware.RequirePending(jwtSecret)` — valid `pending` token required (TOTP/backup-code step only)
- `middleware.RequireCron(cronSecret)` — exact `Authorization: Bearer <CRON_SECRET>` match

A route registered outside one of these groups in `routes.go` is a security bug.

### Request Validation

Gin's `ShouldBindJSON` + struct `binding` tags validate required fields before any DB call in every
handler (e.g. `binding:"required"` on `Title`, `Email`, `EncryptedData`, etc.). There is no separate
Zod-equivalent layer server-side since this is Go, not TypeScript — validation is inline per handler.
(The frontend will use Zod for form validation once built — see `docs/tech-stack.md`.)

### MongoDB Safety

The native `mongo-driver/v2` uses typed BSON query objects — there is no SQL/NoSQL injection
equivalent from string concatenation. Every `:id` route param is parsed with
`bson.ObjectIDFromHex()` and rejected with `400` if malformed, before it reaches a query.

### CORS

Enforced by `internal/middleware/cors.go` against the `ALLOWED_ORIGINS` env var. Allowed origins
today:
- `http://localhost:3000` (dev) / `https://reliva.vercel.app` (prod) — web frontend
- `chrome-extension://<extension-id>` — added once the extension has a stable ID (Phase 6)
- Capacitor WebView origin — added when mobile ships (Phase 7)

`OPTIONS` preflight requests are answered with `204` and the appropriate headers.

---

## Cron Endpoint Protection

```go
func RequireCron(cronSecret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        if c.GetHeader("Authorization") != "Bearer "+cronSecret {
            c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
            return
        }
        c.Next()
    }
}
```

- `CRON_SECRET` is a long random string, set as an env var wherever `apps/server` is deployed
- There is no Vercel Cron involved — `apps/server` isn't hosted on Vercel. Whatever external
  scheduler is configured (Railway cron, a scheduled GitHub Action, etc.) must send this header
- This endpoint does not use `RequireAuth()` — it has its own distinct secret, unrelated to any
  user's JWT

---

## Bearer Token Security (All Clients)

- Tokens are HS256 JWTs signed with `JWT_SECRET`
- Payload: `{ sub: userID, type: "access" | "pending", iat, exp }` plus standard registered claims
- Access token expiry: 30 days. Pending token expiry: 5 minutes
- Tokens are **non-revocable** individually — they're stateless JWTs with no server-side denylist.
  If one is compromised, the only mitigations are waiting for expiry or rotating `JWT_SECRET`
  (which invalidates *every* outstanding token for *every* account, not just the compromised one)
- Never log tokens, in the Go server or any client

---

## Transport Security

- Production traffic should be HTTPS end-to-end: Vercel enforces this for `apps/web`; whatever
  platform hosts `apps/server` (Railway/Fly.io) should have TLS enabled at the edge
- MongoDB Atlas enforces TLS on all connections in production; local Docker MongoDB does not (not a
  concern — it's loopback-only, `27017` is not exposed beyond `localhost` in the default compose
  setup's intent, though the port mapping does bind to the host)
- VAPID push notifications (Phase 5) are signed — push services verify server identity
- Capacitor app (Phase 7) should use HTTPS for all API calls

---

## Sensitive Files Checklist

| File | Status | Notes |
|---|---|---|
| `apps/server/.env` | Gitignored | Contains `JWT_SECRET`, `MONGO_URI`, `CRON_SECRET` |
| `apps/web/.env.local` | Gitignored | Contains only `NEXT_PUBLIC_API_URL` — not sensitive, gitignored by convention anyway |
| `infra/.env` | Gitignored | Contains Mongo root password, `JWT_SECRET`, `CRON_SECRET`, `SEED_PASSWORD` |
| `*.env.example` (all three) | Committed | Keys with empty/placeholder values only |
| `bun.lock` / `go.sum` | Committed | Not sensitive |
| `apps/server/internal/handlers/*.go` | Committed | Logic only — secrets come from env via `config.Config` |
| `apps/web/lib/session.ts` | Committed | Cookie helpers only, no secrets |
| `scripts/seed.ts` / `scripts/seed.cjs` | Committed | Legacy, reads secrets from env — superseded by `apps/server/cmd/seed` |
