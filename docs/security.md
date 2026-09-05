# Security

Security is a first-class concern in Redandan — especially for the password vault module.
This document covers the threat model, the mitigations, and the rules that must not be broken.

---

## Threat Model

Redandan is a personal app with one user. The relevant threats are:

| Threat | Likelihood | Impact |
|---|---|---|
| Unauthorized access to the web app | Medium | High — all data exposed |
| Database breach (MongoDB Atlas) | Low | Critical for vault if not encrypted |
| API abuse via the browser extension | Low | High |
| Session token theft | Low | High |
| Cron endpoint triggered by attacker | Low | Low (only sends notifications) |
| Weak master password for vault | Medium (user error) | Critical — vault decryptable |

---

## Authentication

**How it works:**
- Auth.js (NextAuth v5) manages session lifecycle
- Sessions are JWT-based (signed, not encrypted by default — the secret is `AUTH_SECRET`)
- Credential provider compares submitted password against the bcrypt hash in MongoDB
- Middleware runs at the Vercel Edge and redirects unauthenticated requests to `/login`

**Rules:**
- `AUTH_SECRET` must be a long, random string (32+ characters). Generate with: `openssl rand -base64 32`
- The seeded password must be strong — use a password manager (Redandan itself, eventually)
- Never commit `.env.local` or expose `AUTH_SECRET` in client-side code

**Session expiry:**
- Sessions expire after 30 days by default (Auth.js default)
- Extension tokens follow the same JWT lifetime

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
4. **The browser extension must implement the same decryption logic** — never send the key to the extension from the web app.
5. **If the master password is lost, the vault data is unrecoverable.** This is by design. Document the master password in a secure offline location.

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

**Session validation:**
Every API route handler begins with:
```ts
const session = await auth()
if (!session) return Response.json({ error: 'Unauthorized' }, { status: 401 })
```

**Zod validation:**
Every POST/PATCH request body is parsed through a Zod schema before touching the database.
Invalid input is rejected with a `400` before any DB operation.

**MongoDB injection prevention:**
The native MongoDB driver uses typed query objects — it does not interpolate strings into queries. There is no equivalent of SQL injection via query strings. However:
- Never use `eval` or dynamic key construction with user input
- Always use `new ObjectId(id)` with a try/catch to handle invalid IDs

**CORS:**
Next.js API routes restrict access to same-origin by default.
The browser extension is the only external client — it is whitelisted via `CORS` headers on credentials routes.

---

## Cron Endpoint Protection

The `/api/cron/notify` endpoint must not be callable by anyone except Vercel's cron system.

Protection:
```ts
const authHeader = request.headers.get('authorization')
if (authHeader !== `Bearer ${process.env.CRON_SECRET}`) {
  return Response.json({ error: 'Forbidden' }, { status: 403 })
}
```

- `CRON_SECRET` is a long random string set in Vercel's environment variables
- Vercel Cron automatically sends this header when configured in `vercel.json`

---

## Extension Token Security

The browser extension stores the Auth.js session token in `chrome.storage.local`.

Rules:
- The token is a short-lived JWT (respect the session expiry)
- Extension popup must re-authenticate if the token is expired
- Never expose the token in the extension's UI or logs
- The extension should only request credentials from the API — never write or delete without explicit user action

---

## Transport Security

- All traffic is HTTPS — Vercel enforces this (HTTP redirects to HTTPS)
- MongoDB Atlas enforces TLS on all connections
- VAPID push notifications are signed — push services verify the server identity

---

## Sensitive Files Checklist

| File | Status | Notes |
|---|---|---|
| `.env.local` | Gitignored | Contains all secrets |
| `.env.example` | Committed | Contains keys with empty values only |
| `bun.lock` | Committed | Not sensitive |
| `lib/crypto.ts` | Committed | Contains algorithm, not keys |
| `scripts/seed.ts` | Committed | Reads secrets from env, not hardcoded |
