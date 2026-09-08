# Reliva — Project Planning Document

> Personal Centralized Life & Tech Hub
> Single-user. Personal. Full-fledged.

---

## 1. Vision Summary

Reliva is a personal operating system for your life. It is not a generic SaaS product — it is built specifically for one user (you), solving real friction points:

- You have multiple jobs and lose track of tasks and deadlines
- You want your work life and personal life integrated, not separated
- You waste mental energy remembering passwords
- You want a single hub that notifies you of what matters: a deadline for Job 1, a payroll date for Job 2, a vacation next week

The philosophy: **reduce cognitive load by centralizing everything that matters to you into one trusted system.**

---

## 2. Platform Strategy

The Go server is the **standalone API** for all clients. Every platform is a consumer of the same API — none of them contains business logic or database access.

```
                         ┌──────────────────────────┐
                         │    MongoDB (Atlas/Docker)  │
                         └────────────┬─────────────┘
                                      │
                         ┌────────────▼─────────────┐
                         │   Go + Gin Server         │
                         │   apps/server             │
                         │   (Railway / Fly.io)      │
                         │   ALL business logic      │
                         │   ALL API endpoints       │
                         │   Auth · TOTP · CRUD      │
                         └──┬──────────┬─────────────┘
                            │          │
          ┌─────────────────▼──┐  ┌────▼──────────────────┐
          │  Next.js (Vercel)  │  │  Browser Extension     │
          │  apps/web          │  │  apps/extension        │
          │  FRONTEND ONLY     │  │  Manifest V3           │
          │  Calls Go API      │  │  Calls Go API          │
          └────────┬───────────┘  └───────────────────────┘
                   │
          ┌────────▼───────────┐
          │  Mobile (Phase 7)  │
          │  Capacitor / APK   │
          │  Calls Go API      │
          └────────────────────┘
```

| Platform | Role |
|---|---|
| `apps/server` | Go + Gin — all API, all business logic, all DB access |
| `apps/web` | Next.js — pure frontend, calls Go server API |
| `apps/extension` | Browser extension — calls Go server API |
| Mobile (Phase 7) | Capacitor — calls Go server API |

**Rule: no platform except `apps/server` ever connects to MongoDB or contains business logic.**

---

## 3. Tech Stack

### Backend (`apps/server`)
| Layer | Technology | Reason |
|---|---|---|
| Language | Go 1.23 | Performance, type safety, strong standard library |
| Framework | Gin v1.12 | Fast HTTP router, minimal overhead |
| Database | MongoDB native driver v2 | No ORM, flexible documents |
| Auth | Custom JWT (golang-jwt) + TOTP (pquerna/otp) | Full control, two-step mandatory |
| Crypto | bcrypt (golang.org/x/crypto) | Password hashing |
| Deployment | Railway or Fly.io | Persistent Go process (Vercel doesn't support Go) |

### Frontend (`apps/web`)
| Layer | Technology | Reason |
|---|---|---|
| Framework | Next.js 16 (App Router) | React SSR, file-based routing |
| Language | TypeScript | Type safety |
| UI | shadcn/ui + Tailwind CSS v4 | Already installed |
| Auth (web only) | Cookie-based token storage | Stores Go server JWT in cookie, middleware reads it |
| Vault crypto | Web Crypto API (built-in) | Client-side AES-GCM — server never decrypts |
| Push | Web Push (VAPID) | Browser notifications — Phase 5 |
| Mobile | Capacitor | Wraps Next.js for Android — Phase 7 |
| Deployment | Vercel | Zero-config Next.js |

---

## 4. Core Modules

### 4.1 Authentication

Two-step mandatory login, handled entirely by the Go server:

1. `POST /api/v1/auth/login` — bcrypt verify → pending JWT (5 min)
2. `POST /api/v1/auth/totp/validate` — TOTP verify → access JWT (30 days)

The web frontend:
- Calls the Go server for both steps
- On success, stores the access JWT in a browser cookie
- Next.js middleware reads the cookie to protect routes
- All subsequent API calls send `Authorization: Bearer <token>` to the Go server

TOTP is mandatory. First login (before TOTP setup) redirects to `/settings` where the user completes TOTP setup via the Go server's setup endpoints.

### 4.2 Task Tracker

The core feature. Tasks are the atomic unit of Reliva.

**Fields per task:**
- Title, description
- Context (linked to a job/life area)
- Priority: low / medium / high / urgent
- Status: todo / in_progress / done / archived
- Deadline (optional)
- Reminder date (optional)
- Recurrence (optional: daily, weekly, monthly)
- Tags (free-form)
- Notes

**Contexts / Categories:**
- Work (Job 1, Job 2, etc.)
- Personal
- Health
- Finance
- Travel
- Custom

### 4.3 Calendar & Events

Tracks things that happen at a specific date/time:
- Payroll dates (recurring)
- Project deadlines
- Vacations / travel
- Appointments
- Custom reminders

Tasks with deadlines appear on the calendar automatically.

### 4.4 Password Manager

Encrypted credential vault — the most security-sensitive module.

**Architecture:**
- Master password is **never sent to any server**
- A `derivedKey` is generated client-side from master password using PBKDF2 (Web Crypto API)
- Each credential is encrypted client-side with AES-GCM before storage
- Go server stores only ciphertext + IV + salt — server cannot decrypt
- Decryption happens entirely in the browser/extension

### 4.5 Notifications & Reminders

- Web Push notifications (VAPID) — Phase 5
- Capacitor Push Plugin for mobile — Phase 7
- Cron job hits the Go server's `POST /api/v1/cron/notify` daily

### 4.6 Dashboard (Home)

Unified view:
- Tasks due today and this week, grouped by context
- Upcoming calendar events (next 7 days)
- Quick-add task input
- Overdue items highlighted

---

## 5. Database Design (MongoDB)

No ORM. All queries use the native Go `mongo-driver` directly in `apps/server`.

### Collections

#### `users`
```json
{
  "_id": ObjectId,
  "email": "string",
  "passwordHash": "string",
  "totpSecret": "string",
  "totpEnabled": false,
  "totpPendingSecret": "string",
  "backupCodes": ["string (bcrypt hash)"],
  "createdAt": "Date"
}
```

#### `contexts`
```json
{
  "_id": ObjectId,
  "userId": ObjectId,
  "name": "Job 1 — Company Name",
  "slug": "job-1",
  "color": "#hexcode",
  "icon": "string",
  "type": "work | personal | health | finance | travel | custom",
  "order": 0,
  "createdAt": "Date",
  "updatedAt": "Date"
}
```

#### `tasks`
```json
{
  "_id": ObjectId,
  "userId": ObjectId,
  "contextId": ObjectId,
  "title": "string",
  "description": "string",
  "priority": "low | medium | high | urgent",
  "status": "todo | in_progress | done | archived",
  "dueDate": "Date | null",
  "reminderAt": "Date | null",
  "recurrence": "none | daily | weekly | monthly",
  "tags": ["string"],
  "notes": "string",
  "completedAt": "Date | null",
  "createdAt": "Date",
  "updatedAt": "Date"
}
```

#### `events`
```json
{
  "_id": ObjectId,
  "userId": ObjectId,
  "contextId": "ObjectId | null",
  "title": "string",
  "startTime": "Date",
  "endTime": "Date | null",
  "allDay": false,
  "recurrence": "none | monthly | annually",
  "notes": "string",
  "createdAt": "Date",
  "updatedAt": "Date"
}
```

#### `credentials`
```json
{
  "_id": ObjectId,
  "userId": ObjectId,
  "site": "string",
  "siteUrl": "string",
  "username": "string",
  "encryptedData": "string (base64url)",
  "iv": "string (base64url)",
  "salt": "string (base64url)",
  "encryptedNotes": "string | null",
  "notesIv": "string | null",
  "tags": ["string"],
  "createdAt": "Date",
  "updatedAt": "Date"
}
```

#### `notifications`
```json
{
  "_id": ObjectId,
  "userId": ObjectId,
  "title": "string",
  "body": "string",
  "refId": "ObjectId | null",
  "refType": "task | event | null",
  "read": false,
  "createdAt": "Date"
}
```

#### `push_subscriptions`
```json
{
  "_id": ObjectId,
  "userId": ObjectId,
  "endpoint": "string",
  "p256dh": "string",
  "auth": "string",
  "createdAt": "Date"
}
```

---

## 6. API Reference

All API endpoints live in `apps/server` (Go). See `openapi.yaml` at the monorepo root for the full contract.

Base URL:
- Local: `http://localhost:8080`
- Production: `https://<your-server>.railway.app`

Authentication:
- Step 1: `POST /api/v1/auth/login` → pending token
- Step 2: `POST /api/v1/auth/totp/validate` → access token
- Send `Authorization: Bearer <access_token>` on all protected endpoints

The web frontend calls these endpoints directly from server components (using the stored cookie token) and from client components (via `fetch`).

---

## 7. Frontend Page Structure (`apps/web`)

```
app/
  layout.tsx              — Root layout, fonts, theme provider
  (auth)/
    login/page.tsx        — Two-step login (password → TOTP)
  (app)/
    layout.tsx            — Auth guard: reads cookie, redirects if missing or TOTP not set up
    page.tsx              — Dashboard
    tasks/
      page.tsx            — Task list
      new/page.tsx        — Create task
      [id]/page.tsx       — Task detail / edit
    calendar/page.tsx     — Calendar view
    vault/
      page.tsx            — Credential list
      new/page.tsx        — Add credential
      [id]/page.tsx       — View/edit credential
    settings/page.tsx     — TOTP setup, account settings

lib/
  token.ts                — Cookie read/write helpers for the Go server JWT
  api.ts                  — Typed fetch wrapper for Go server endpoints
  crypto.ts               — Client-side AES-GCM encryption (vault only)
  types.gen.ts            — Auto-generated from openapi.yaml (do not edit)
```

`apps/web` has **no API routes** and **no database connection**. All data comes from `apps/server`.

---

## 8. Browser Extension (`apps/extension`)

```
extension/
  manifest.json           — Manifest V3 (Chrome + Firefox)
  popup/
    index.html
    popup.ts              — Credential lookup, task quick-add
  background/
    service-worker.ts     — Token storage, badge count
  content/
    autofill.ts           — DOM injection for login form autofill
```

Stores JWT in `chrome.storage.local`. Calls Go server directly.

---

## 9. Mobile Strategy (Phase 7)

After the web app is stable:

1. Ensure all pages use client-side data fetching (no SSR-only patterns)
2. Add Capacitor: `bun cap init reliva com.personal.reliva`
3. Configure `capacitor.config.ts` to point to the production Go server URL
4. `bun cap add android`
5. Use Capacitor plugins for push notifications, haptics, status bar

---

## 10. Security

| Concern | Solution |
|---|---|
| All API access | JWT bearer token issued by Go server, required on every protected endpoint |
| Two-factor auth | TOTP mandatory — no bypass, no skip, no disable |
| Password storage | bcrypt (Go server), never stored plaintext |
| Vault encryption | AES-GCM client-side (Web Crypto API) — server stores only ciphertext |
| Token storage (web) | Cookie (`reliva_token`); not httpOnly since no server sets it, but scoped to the domain |
| Token storage (extension) | `chrome.storage.local` |
| Transport | HTTPS only in production |
| MongoDB injection | Parameterized BSON queries — no string interpolation |
| CORS | Go server restricts to `ALLOWED_ORIGINS` env var |

---

## 11. Development Roadmap

### Phase 1 — Foundation ✅ (In Progress)
- [x] Go server scaffold — all auth + CRUD endpoints
- [x] Docker Compose infra (MongoDB + server)
- [x] OpenAPI spec (`openapi.yaml`) + TypeScript type generation
- [ ] Next.js login page — calls Go server auth endpoints
- [ ] Token cookie management in Next.js (`lib/token.ts`)
- [ ] Next.js middleware — reads cookie, protects routes
- [ ] Settings page — TOTP setup via Go server endpoints
- [ ] Dashboard placeholder
- [ ] `apps/server` seed script (Go or separate tool) to create the admin user

### Phase 2 — Task Tracker
- [ ] Task list page — fetch from Go server
- [ ] Create/edit task forms
- [ ] Context sidebar
- [ ] Dashboard: today's tasks + upcoming
- [ ] Filtering by context, status, priority

### Phase 3 — Calendar & Events
- [ ] Calendar view (month + agenda)
- [ ] Event create/edit
- [ ] Tasks with deadlines on calendar
- [ ] Recurring event support

### Phase 4 — Password Vault
- [ ] `lib/crypto.ts` — PBKDF2 + AES-GCM helpers
- [ ] Vault list page
- [ ] Add/view credential (encrypt before POST, decrypt on demand)
- [ ] Master password unlock UI

### Phase 5 — Notifications
- [ ] VAPID key generation
- [ ] Push subscription registration
- [ ] Service worker
- [ ] Notification bell in nav

### Phase 6 — Browser Extension
- [ ] Manifest V3 scaffold
- [ ] Two-step auth in extension popup
- [ ] Autofill content script
- [ ] Quick task-add

### Phase 7 — Mobile (Android)
- [ ] Capacitor setup
- [ ] Android build + APK
- [ ] Native push notifications

---

## 12. Environment Variables

### `apps/server` (Go)
```env
MONGO_URI=
DB_NAME=reliva
JWT_SECRET=
SERVER_PORT=8080
ALLOWED_ORIGINS=http://localhost:3000
CRON_SECRET=
VAPID_PUBLIC_KEY=
VAPID_PRIVATE_KEY=
VAPID_EMAIL=
```

### `apps/web` (Next.js)
```env
# URL of the Go server
NEXT_PUBLIC_API_URL=http://localhost:8080
```

---

*This document is the single source of truth for Reliva's architecture and roadmap. Update it as the project evolves.*
