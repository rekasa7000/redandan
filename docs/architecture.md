# Architecture

## Core Principle: Standalone Server

The API is the single source of truth. It has no knowledge of which client is consuming it.
The web frontend, mobile app, and browser extension are all equal clients — none receives special treatment.

This means:
- All data operations go through `/api/**` REST endpoints
- The server never returns HTML or client-specific responses
- Adding a new client (CLI tool, desktop app, another mobile platform) requires zero changes to the server
- The web frontend is a React app that calls the API, just like any other client

---

## System Diagram

```
                    ┌──────────────────────────────┐
                    │      MongoDB Atlas (Cloud)    │
                    │  users / tasks / contexts     │
                    │  events / credentials         │
                    │  notifications                │
                    └──────────────┬───────────────┘
                                   │ native mongodb driver
                    ┌──────────────▼───────────────┐
                    │                              │
                    │     Redandan API (Vercel)    │
                    │       /app/api/**            │
                    │                              │
                    │  Auth: Session cookie (web)  │
                    │        Bearer token (others) │
                    │  2FA:  TOTP required always  │
                    │                              │
                    └──────┬────────────┬──────────┘
                           │            │
           ┌───────────────▼──┐   ┌─────▼────────────────────────┐
           │  Web Frontend    │   │  Bearer Token Clients         │
           │  (Next.js React) │   │                               │
           │                  │   │  ┌─────────────────────────┐  │
           │  Cookie session  │   │  │  Browser Extension      │  │
           │  Same origin     │   │  │  (Manifest V3)          │  │
           │  Server components│  │  │  chrome.storage.local   │  │
           └──────────────────┘   │  └─────────────────────────┘  │
                                  │                               │
                                  │  ┌─────────────────────────┐  │
                                  │  │  Mobile App             │  │
                                  │  │  (Capacitor/Android)    │  │
                                  │  │  Secure token storage   │  │
                                  │  └─────────────────────────┘  │
                                  └───────────────────────────────┘
```

---

## Layers

### 1. Database Layer — MongoDB Atlas
- Cloud-hosted MongoDB (free M0 tier)
- Connected via the official `mongodb` native driver
- Connection is pooled via a singleton in `lib/db.ts`
- No ORM. All queries are written directly in TypeScript

### 2. API Layer — Next.js Route Handlers (Standalone)
- Lives in `app/api/`
- Every route validates the caller's identity before any logic runs
- Accepts two authentication methods: session cookie (web) or bearer token (mobile/extension)
- Returns pure JSON. No HTML. No redirects.
- CORS headers configured to allow requests from trusted non-same-origin clients
- This layer is client-agnostic — it does not know or care which frontend is calling it

### 3. Auth Layer — Two-Factor: Password + TOTP
- **Step 1:** Username + password (bcrypt comparison)
- **Step 2:** 6-digit TOTP code from an authenticator app (Google Authenticator, Authy, etc.)
- Both factors must pass before any session or token is issued
- TOTP is validated server-side using `otplib` against the stored secret
- The TOTP secret is stored encrypted in the `users` collection

Auth.js (NextAuth v5) manages the session lifecycle for the web client.
A separate `POST /api/auth/token` endpoint issues signed JWTs for non-cookie clients (extension, mobile).

### 4. Web Frontend Layer — React (App Router)
- One of three clients of the API — not privileged over others
- Server Components can call the DB directly as an internal optimization for the web (RSC pattern)
- Client Components always call the API via fetch
- Auth.js session available via `auth()` in server components and `useSession()` in client components
- `middleware.ts` guards all `(app)` routes at the edge

### 5. Extension Layer
- Authenticates with the API using a stored bearer token (`chrome.storage.local`)
- Performs TOTP during its own login flow (popup → API → token issued)
- Content script detects login forms and autofills from decrypted vault credentials
- Popup provides credential search and quick task creation

### 6. Mobile Layer (Phase 7)
- Capacitor wraps the Next.js static export
- Authenticates using a bearer token stored in Capacitor's secure storage
- Native push notification plugins replace web push
- Calls the same API endpoints as the extension

---

## Authentication Flow (All Clients)

```
Client submits: username + password
  → Server: bcrypt verify password
  → If invalid: 401
  → If valid: prompt for TOTP

Client submits: 6-digit TOTP code
  → Server: otplib.totp.verify(code, secret)
  → If invalid: 401 (with attempt counter)
  → If valid:
      Web client  → issue Auth.js session cookie (HttpOnly, Secure)
      Other client → issue signed JWT bearer token (POST /api/auth/token)

All subsequent requests:
  → Web: sends session cookie automatically (browser behavior)
  → Extension/Mobile: sends Authorization: Bearer <token> header
```

---

## Data Flow Examples

### Task Creation (Web Frontend)
```
User submits task form
  → Client Component validates locally (Zod)
  → POST /api/tasks (cookie sent automatically)
  → Route Handler: validates session cookie via auth()
  → Zod schema validates request body
  → Inserts into MongoDB tasks collection
  → Returns 201 with created task document
  → UI updates
```

### Task Creation (Mobile App)
```
User submits task form in Capacitor app
  → POST /api/tasks (Authorization: Bearer <token>)
  → Route Handler: validates bearer token via lib/auth.ts validateToken()
  → Zod schema validates request body
  → Inserts into MongoDB tasks collection
  → Returns 201 with created task document
  → UI updates
```
*The route handler logic is identical — only the auth validation method differs.*

### Password Autofill (Extension)
```
User visits a login page
  → Content script detects <input type="password">
  → Reads current tab domain
  → GET /api/credentials?site=github.com (Authorization: Bearer <token>)
  → Server returns encrypted ciphertext
  → Popup decrypts client-side using master key (Web Crypto API)
  → Injects plaintext into form inputs
```

### Scheduled Notification
```
Vercel Cron Job fires at 8:00 AM
  → GET /api/cron/notify (Authorization: Bearer <CRON_SECRET>)
  → Route Handler: validates CRON_SECRET header
  → Queries tasks due today + overdue
  → Queries events within 2 days
  → Sends Web Push to all stored subscriptions
  → Logs to notifications collection
```

---

## CORS Policy

The API must be reachable from non-same-origin clients.

| Client | Origin | Method |
|---|---|---|
| Web frontend | Same origin | Session cookie |
| Browser extension | `chrome-extension://<id>` | Bearer token |
| Mobile (Capacitor) | Capacitor WebView / native | Bearer token |

CORS headers on all `/api/**` routes:
```
Access-Control-Allow-Origin: <whitelisted origins>
Access-Control-Allow-Methods: GET, POST, PATCH, DELETE, OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization
```

The allowed origins list is managed via `ALLOWED_ORIGINS` environment variable.
The extension's origin (`chrome-extension://...`) is added after the extension is built and has a stable ID.

---

## Key Design Constraints

| Constraint | Rationale |
|---|---|
| Standalone API | Any client can be built without touching the server |
| TOTP on all clients | Maximum security — single factor alone is not enough |
| Dual auth method | Cookies for web (browser handles them); bearer tokens for extension/mobile |
| No ORM | Direct MongoDB queries, full control |
| Client-side crypto | Vault passwords never decryptable by the server |
| One repo | Web + API in one Next.js project; extension in a subfolder |

---

## Deployment Architecture

```
GitHub (main branch)
  → Vercel auto-deploys on push
  → Environment variables set in Vercel dashboard
  → Vercel Cron Jobs configured in vercel.json
  → MongoDB Atlas accessed via MONGODB_URI env var
```

The extension is built separately (`bun run build:extension`) and loaded unpacked in the browser.
The mobile APK is built locally using Android Studio + Capacitor and sideloaded.
