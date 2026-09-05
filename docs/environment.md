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
| `AUTH_SECRET` | Yes | Signs JWT session tokens. Must be long and random. |
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

---

### Push Notifications (VAPID)

| Variable | Required | Description |
|---|---|---|
| `VAPID_PUBLIC_KEY` | Phase 5 | Public key sent to browsers for push subscription |
| `VAPID_PRIVATE_KEY` | Phase 5 | Private key used to sign push messages |
| `VAPID_SUBJECT` | Phase 5 | Contact email for push service providers |

**Generate VAPID keys (one-time):**
```bash
npx web-push generate-vapid-keys
```

**Values:**
```
VAPID_PUBLIC_KEY=<base64 public key>
VAPID_PRIVATE_KEY=<base64 private key>
VAPID_SUBJECT=mailto:your-email@example.com
```

`VAPID_PUBLIC_KEY` is also used client-side (it is safe to expose):
```ts
// In your push subscription code
const PUBLIC_KEY = process.env.NEXT_PUBLIC_VAPID_PUBLIC_KEY!
```

Note: prefix with `NEXT_PUBLIC_` to make it accessible in the browser.

---

### Cron Security

| Variable | Required | Description |
|---|---|---|
| `CRON_SECRET` | Phase 5 | Protects the cron endpoint from unauthorized calls |

**Generate:**
```bash
openssl rand -base64 32
```

**Value:**
```
CRON_SECRET=<random string>
```

Vercel Cron Jobs send this automatically via the `Authorization: Bearer <CRON_SECRET>` header when configured in `vercel.json`.

---

### Seed Script (Local Only)

| Variable | Required | Description |
|---|---|---|
| `SEED_USERNAME` | Setup only | Username for the admin user (you) |
| `SEED_PASSWORD` | Setup only | Password for the admin user (you) |

These are only used when running `bun run scripts/seed.ts` locally.
They do not need to be set in Vercel's environment.

```
SEED_USERNAME=your-username
SEED_PASSWORD=your-strong-password
```

---

## `.env.example`

This file is committed to the repository with all keys present but values empty:

```env
# MongoDB
MONGODB_URI=

# Auth.js
AUTH_SECRET=
AUTH_URL=

# Push Notifications (VAPID) — Phase 5
NEXT_PUBLIC_VAPID_PUBLIC_KEY=
VAPID_PRIVATE_KEY=
VAPID_SUBJECT=

# Cron Security — Phase 5
CRON_SECRET=

# Seed Script (local only)
SEED_USERNAME=
SEED_PASSWORD=
```

---

## Vercel Setup Instructions

1. Go to your Vercel project → Settings → Environment Variables
2. Add each variable from the table above
3. Set the environment: `Production` and `Preview` (not needed for `SEED_*`)
4. Redeploy after adding new variables

---

## Local Development

Copy `.env.example` to `.env.local`:
```bash
cp .env.example .env.local
```

Fill in the values. Then:
```bash
bun dev
```

The app reads `.env.local` automatically in development.
