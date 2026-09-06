# Security

Security is a first-class concern in Redandan — especially for the password vault and authentication.
This document covers the threat model, the mitigations, and the rules that must not be broken.

---

## Threat Model

Redandan is a personal app with one user. The relevant threats are:

| Threat | Likelihood | Impact |
|---|---|---|
| Unauthorized access to the web app | Medium | High — all data exposed |
| Password stolen or guessed | Low-Medium | Mitigated by TOTP (single factor not enough) |
| TOTP code intercepted or stolen | Very Low | High — valid only for 30 seconds |
| Database breach (MongoDB Atlas) | Low | Critical for vault if not encrypted |
| Bearer token stolen (extension/mobile) | Low | High — full API access |
| Session cookie stolen | Low | High — full web access |
| Cron endpoint triggered by attacker | Low | Low (only sends notifications) |
| Weak master password for vault | Medium (user error) | Critical — vault decryptable |
| Authenticator app + backup codes both lost | Low (user error) | Account lockout |

---

## Authentication — Two-Factor (Password + TOTP)

Authentication requires **two independent factors** on every login. Passing one does not grant access.

### Step 1 — Password

- The stored password is hashed using **bcrypt** (cost factor 12) in the `users` collection
- On login, the submitted password is compared with `bcrypt.compare()`
- A failed password check returns `401` with no information about which factor failed
- After 5 failed attempts within 15 minutes, the login endpoint rate-limits the IP (implementation: sliding window counter in MongoDB or Upstash Redis)

### Step 2 — TOTP (Time-Based One-Time Password)

- The user opens their authenticator app (Google Authenticator, Authy, etc.) and enters the 6-digit code
- The server validates using `otplib.authenticator.verify({ token, secret })`
- TOTP codes are valid for ±1 time step (30 seconds) to account for clock drift
- A valid TOTP code from the current window is accepted once only (replay prevention via a "used token" cache)
- After 5 failed TOTP attempts, the same rate limit applies as for passwords

### Session / Token Issuance

After both factors pass:

| Client | Auth artifact | Storage | Expiry |
|---|---|---|---|
| Web browser | Auth.js session cookie (HttpOnly, Secure, SameSite=Lax) | Browser cookie jar | 30 days |
| Browser extension | Signed JWT bearer token | `chrome.storage.local` | 30 days |
| Mobile app | Signed JWT bearer token | Capacitor SecureStorage | 30 days |

All tokens are signed with `AUTH_SECRET`. They are not encrypted — the payload is readable but tamper-proof.

### TOTP Setup (One-Time)

1. A random 20-byte TOTP secret is generated server-side using `otplib.authenticator.generateSecret()`
2. The secret is stored encrypted in the `users` collection (AES-256 with a key derived from `AUTH_SECRET`)
3. A QR code is generated using the `qrcode` library and displayed in the settings page
4. The user scans the QR code with their authenticator app
5. The user enters a verification code to confirm setup is working
6. 10 single-use backup codes are generated and shown once — the user must save them offline

### Backup Codes

- Generated at TOTP setup time using `crypto.randomBytes()`
- Each code is 10 characters, alphanumeric, formatted as `XXXXX-XXXXX`
- Stored as bcrypt hashes in the `users.backupCodes` array (plaintext never stored)
- Using a backup code marks it as consumed — it cannot be used again
- After all backup codes are exhausted, new ones can be generated (requires TOTP to do so)
- If the authenticator app and all backup codes are lost, recovery requires direct database access

**Rules:**
- `AUTH_SECRET` must be a long, random string (32+ characters). Generate: `openssl rand -base64 32`
- The seeded password must be strong and unique
- Never commit `.env.local` or expose `AUTH_SECRET` in client-side code
- Store backup codes offline — printed or in a physically secure location

---

## Password Vault — Zero-Knowledge Design

This is the most security-critical module. The design ensures the server has zero ability to decrypt vault contents.

### Encryption Flow (Client → Server)

```
User types master password
  → PBKDF2 key derivation (100,000 iterations, SHA-256, random 16-byte salt)
  → Produces a CryptoKey (AES-GCM 256-bit)
  → Kept in memory only (never stored, never sent)

User adds a credential
  → Generate random 12-byte IV
  → AES-GCM encrypt plaintext password with derived key + IV
  → POST /api/credentials with { encryptedPassword, iv, salt, username, site }

Server stores verbatim
  → MongoDB receives: { encryptedPassword, iv, salt }
  → Server never sees plaintext, never derives the key
```

### Decryption Flow (Server → Client)

```
User opens vault
  → Prompted for master password (re-derives key in memory)
  → GET /api/credentials returns ciphertext + iv + salt
  → Client decrypts each entry with the derived key
  → Plaintext displayed in browser memory only
```

### Key Rules — Never Break These

1. **Never add server-side decryption logic.** The API returns ciphertext and that's it.
2. **Never log credential plaintext** anywhere — not in API routes, not in the browser console.
3. **Never store the master password or derived key in localStorage or a cookie.** Memory only.
4. **The browser extension must implement the same decryption logic** — never send the key from the web app to the extension.
5. **If the master password is lost, the vault data is unrecoverable.** This is by design. Store the master password in a secure offline location.

### Cryptographic Parameters

| Parameter | Value |
|---|---|
| Key derivation | PBKDF2 |
| Hash | SHA-256 |
| Iterations | 100,000 |
| Salt length | 16 bytes (random per credential) |
| Encryption | AES-GCM |
| Key length | 256 bits |
| IV length | 12 bytes (random per credential) |

These parameters are defined in `lib/crypto.ts` and must not be changed without migrating existing data.

---

## API Security

### Caller Validation

Every API route handler calls `validateCaller(request)` from `lib/auth.ts` before any other logic.
This function accepts either a session cookie or a bearer token:

```ts
// lib/auth.ts
export async function validateCaller(request: Request) {
  // Try bearer token first (extension, mobile)
  const authHeader = request.headers.get('authorization')
  if (authHeader?.startsWith('Bearer ')) {
    const token = authHeader.slice(7)
    return verifyJwt(token) // throws if invalid
  }
  // Fall back to session cookie (web)
  const session = await auth()
  if (!session) throw new Error('Unauthorized')
  return session
}
```

A route handler that does not call `validateCaller()` is a security bug.

### Zod Validation

Every POST/PATCH request body is parsed through a Zod schema before touching the database.
Invalid input is rejected with a `400` before any DB operation runs.

### MongoDB Safety

The native MongoDB driver uses typed query objects — there is no SQL injection equivalent.
However:
- Never use `eval` or dynamic key construction with user input
- Always use `new ObjectId(id)` wrapped in a try/catch — invalid ObjectId strings throw

### CORS

CORS headers are set on all `/api/**` routes. Allowed origins:
- Same origin (web frontend, implicit)
- `chrome-extension://<extension-id>` (added after extension build)
- Capacitor WebView origin (configured in `ALLOWED_ORIGINS` env var)

Preflight (`OPTIONS`) requests are handled and return the correct headers.

---

## Cron Endpoint Protection

```ts
const authHeader = request.headers.get('authorization')
if (authHeader !== `Bearer ${process.env.CRON_SECRET}`) {
  return Response.json({ error: 'Forbidden' }, { status: 403 })
}
```

- `CRON_SECRET` is a long random string set in Vercel's environment variables
- Vercel Cron sends this header automatically when configured in `vercel.json`
- This endpoint does not use `validateCaller()` — it uses its own distinct secret

---

## Bearer Token Security (Extension / Mobile)

- Tokens are signed JWTs using `AUTH_SECRET` as the signing key
- Payload contains only: `{ sub: userId, iat, exp }`
- Expiry: 30 days (same as web session)
- The extension refreshes the token silently via the background service worker before expiry
- Tokens are non-revocable (stateless JWT) — if a token is compromised, wait for expiry or rotate `AUTH_SECRET` (invalidates all tokens)
- Never log tokens in the extension UI or service worker console

---

## Transport Security

- All traffic is HTTPS — Vercel enforces this (HTTP redirects to HTTPS automatically)
- MongoDB Atlas enforces TLS on all connections
- VAPID push notifications are signed — push services verify the server identity
- Capacitor app uses HTTPS for all API calls (`cleartext: false` in `capacitor.config.ts`)

---

## Sensitive Files Checklist

| File | Status | Notes |
|---|---|---|
| `.env.local` | Gitignored | Contains all secrets |
| `.env.example` | Committed | Keys with empty values only |
| `bun.lock` | Committed | Not sensitive |
| `lib/crypto.ts` | Committed | Algorithm constants only, no keys |
| `lib/auth.ts` | Committed | Logic only — secrets come from env |
| `scripts/seed.ts` | Committed | Reads secrets from env, not hardcoded |
