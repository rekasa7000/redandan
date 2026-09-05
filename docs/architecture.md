# Architecture

## System Overview

Redandan is built on a single Next.js codebase that serves as both the frontend and the backend API. All platforms (web, mobile, browser extension) talk to the same API hosted on Vercel.

```
                    ┌──────────────────────────┐
                    │    MongoDB Atlas (Cloud)  │
                    │    - users               │
                    │    - tasks               │
                    │    - contexts            │
                    │    - events              │
                    │    - credentials         │
                    │    - notifications       │
                    └─────────────┬────────────┘
                                  │ native mongodb driver
                    ┌─────────────▼────────────┐
                    │   Next.js on Vercel       │
                    │                          │
                    │   /app/api/**            │  ← REST API (Route Handlers)
                    │   /app/(app)/**          │  ← React Frontend (RSC + Client)
                    │   middleware.ts          │  ← Auth guard on all routes
                    └──────┬──────────┬────────┘
                           │          │
          ┌────────────────▼──┐    ┌──▼────────────────────────┐
          │   Mobile App      │    │   Browser Extension        │
          │   (Capacitor)     │    │   (Manifest V3)            │
          │                   │    │                            │
          │   Wraps web UI    │    │   popup.html (mini UI)     │
          │   Native push     │    │   content.ts (autofill)    │
          │   Android/iOS     │    │   service-worker.ts        │
          └───────────────────┘    └────────────────────────────┘
```

---

## Layers

### 1. Database Layer — MongoDB Atlas
- Cloud-hosted MongoDB (free M0 tier is sufficient for personal use)
- Connected via the official `mongodb` native driver
- Connection is pooled via a singleton in `lib/db.ts`
- No ORM. All queries are written directly in TypeScript

### 2. API Layer — Next.js Route Handlers
- Lives in `app/api/`
- Every route is protected by Auth.js session validation
- Returns JSON. Follows REST conventions
- Handles all CRUD for tasks, events, credentials, notifications, contexts

### 3. Frontend Layer — React (App Router)
- Server Components for data fetching (no waterfall)
- Client Components for interactivity (forms, modals, drag-drop)
- Auth.js session available via `auth()` in server components and `useSession()` in client components
- State managed locally with React state / URL search params — no global state library needed at this scale

### 4. Auth Layer — Auth.js (NextAuth v5)
- Single credential provider (username + password)
- Passwords hashed with bcrypt
- Sessions are JWT-based
- `middleware.ts` redirects unauthenticated users to `/login` for all `(app)` routes

### 5. Extension Layer
- Separate build in `extension/`
- Communicates with Vercel API using a stored auth token (`chrome.storage.local`)
- Content script injects autofill into login forms
- Popup provides credential search and quick task creation

### 6. Mobile Layer (Phase 7)
- Capacitor wraps the Next.js static export
- Native plugins for push notifications and local notifications
- Deployed as an APK (Android) for personal sideloading

---

## Data Flow

### Task Creation (Web)
```
User fills form
  → Client Component validates (Zod)
  → POST /api/tasks
  → Route Handler validates session
  → Inserts into MongoDB tasks collection
  → Returns created task
  → UI optimistically updates
```

### Password Autofill (Extension)
```
User visits login page
  → Content script detects login form
  → Reads current domain
  → GET /api/credentials?site=example.com (with auth token)
  → Server returns encrypted ciphertext
  → Popup decrypts client-side with master key (Web Crypto API)
  → Autofills username + password into form
```

### Notification Delivery
```
Vercel Cron Job fires (e.g. daily at 8am)
  → GET /api/cron/notify (with CRON_SECRET header)
  → Route handler queries tasks due today / overdue
  → Sends Web Push to stored push subscriptions
  → Logs to notifications collection
```

---

## Key Design Constraints

| Constraint | Rationale |
|---|---|
| No ORM | Direct MongoDB queries give full control; no abstraction overhead |
| Single user | No need for multi-tenancy, roles, or org structures |
| API-first | Extension and mobile can both consume the same endpoints |
| Client-side crypto | Vault passwords must never be decryptable by the server |
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

The extension is built separately and packaged manually for browser stores (or loaded unpacked for personal use).

The mobile APK is built locally using Android Studio + Capacitor and sideloaded.
