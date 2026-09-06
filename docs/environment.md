# Environment Variables

All secrets are stored in `.env.local` locally and in Vercel's environment variables dashboard for production.

**Never commit `.env.local` to git.** It is in `.gitignore`.
Use `.env.example` as the committed reference with empty values.

---

## Full Variable Reference

### MongoDB

| Variable | Required | Description |
|---|---|---|
| `MONGODB_URI` | Yes | Full MongoDB Atlas connection string |

**Format:**
```
MONGODB_URI=mongodb+srv://<username>:<password>@<cluster>.mongodb.net/redandan?retryWrites=true&w=majority
```

Get this from MongoDB Atlas → your cluster → Connect → Drivers → Node.js.

---

### Auth.js

| Variable | Required | Description |
|---|---|---|
| `AUTH_SECRET` | Yes | Signs JWT session tokens and bearer tokens. Must be long and random. |
| `AUTH_URL` | Yes | Full URL of the deployed app (no trailing slash) |

**Generate `AUTH_SECRET`:**
```bash
openssl rand -base64 32
```

**Values:**
```
AUTH_SECRET=<32+ character random string>
AUTH_URL=https://redandan.vercel.app
```

In development:
```
AUTH_URL=http://localhost:3000
```

`AUTH_SECRET` is also used to encrypt the TOTP secret stored in MongoDB.
If you rotate `AUTH_SECRET`, all existing sessions and bearer tokens are immediately invalidated,
and the stored TOTP secret will no longer be decryptable — you will need to re-run TOTP setup.

---

### CORS

| Variable | Required | Description |
|---|---|---|
| `ALLOWED_ORIGINS` | Yes | Comma-separated list of allowed cross-origins for the API |

**Value:**
```
ALLOWED_ORIGINS=https://redandan.vercel.app,chrome-extension://<extension-id>
```

In development:
```
ALLOWED_ORIGINS=http://localhost:3000
```

The extension ID is stable once the extension is loaded. Get it from `chrome://extensions` and add it here.
After updating this variable, redeploy.

---

### Push Notifications (VAPID)

| Variable | Required | Description |
|---|---|---|
| `NEXT_PUBLIC_VAPID_PUBLIC_KEY` | Phase 5 | Public key sent to browsers for push subscription |
| `VAPID_PRIVATE_KEY` | Phase 5 | Private key used to sign push messages |
| `VAPID_SUBJECT` | Phase 5 | Contact email for push service providers |

**Generate VAPID keys (one-time):**
```bash
npx web-push generate-vapid-keys
```

**Values:**
```
NEXT_PUBLIC_VAPID_PUBLIC_KEY=<base64 public key>
VAPID_PRIVATE_KEY=<base64 private key>
VAPID_SUBJECT=mailto:your-email@example.com
```

Note the `NEXT_PUBLIC_` prefix on the public key — this makes it accessible in the browser for
constructing the push subscription. The private key must never be exposed client-side.

---

### Cron Security

| Variable | Required | Description |
|---|---|---|
| `CRON_SECRET` | Phase 5 | Protects the `/api/cron/notify` endpoint |

**Generate:**
```bash
openssl rand -base64 32
```

Vercel Cron Jobs send this automatically via `Authorization: Bearer <CRON_SECRET>` when configured in `vercel.json`.

---

### Seed Script (Local Only)

| Variable | Required | Description |
|---|---|---|
| `SEED_USERNAME` | Setup only | Username for the admin user |
| `SEED_PASSWORD` | Setup only | Password for the admin user |

Used only when running `bun run scripts/seed.ts`. Not needed in Vercel's environment.

```
SEED_USERNAME=your-username
SEED_PASSWORD=your-strong-password
```

---

## `.env.example`

```env
# MongoDB
MONGODB_URI=

# Auth.js — also used to sign bearer tokens and encrypt the TOTP secret
AUTH_SECRET=
AUTH_URL=

# CORS — comma-separated allowed origins for the standalone API
ALLOWED_ORIGINS=

# Push Notifications (VAPID) — added in Phase 5
NEXT_PUBLIC_VAPID_PUBLIC_KEY=
VAPID_PRIVATE_KEY=
VAPID_SUBJECT=

# Cron Security — added in Phase 5
CRON_SECRET=

# Seed Script (local only — not needed in Vercel)
SEED_USERNAME=
SEED_PASSWORD=
```

---

## Vercel Setup Instructions

1. Go to your Vercel project → Settings → Environment Variables
2. Add each variable (excluding `SEED_*`) for `Production` and `Preview`
3. Redeploy after adding new variables — environment changes do not take effect until redeployment

---

## Local Development

```bash
cp .env.example .env.local
# fill in all values
bun dev
```

---

## Security Notes

- All secrets should be unique values — never reuse passwords or keys across services
- Rotate `AUTH_SECRET` only intentionally — it invalidates all active sessions and bearer tokens
- The TOTP secret stored in MongoDB is encrypted with `AUTH_SECRET`. Rotating the secret requires re-running TOTP setup
- `VAPID_PRIVATE_KEY` and `CRON_SECRET` should each be independently generated
