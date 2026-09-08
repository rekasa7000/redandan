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

The same Next.js codebase serves all platforms:

```
                          ┌─────────────────────────┐
                          │   MongoDB Atlas (Cloud)  │
                          └────────────┬────────────┘
                                       │
                          ┌────────────▼────────────┐
                          │  Next.js App (Vercel)   │
                          │  - Frontend (React)     │
                          │  - API Route Handlers   │
                          └──┬─────────┬────────────┘
                             │         │
              ┌──────────────▼──┐   ┌──▼──────────────────┐
              │  Mobile App     │   │  Browser Extension   │
              │  (Capacitor)    │   │  (Chrome + Firefox)  │
              │  Android/iOS    │   │  Manifest V3         │
              └─────────────────┘   └──────────────────────┘
```

| Platform | Approach |
|---|---|
| Web | Next.js deployed on Vercel |
| Mobile | Capacitor wraps the built Next.js app |
| Browser Extension | Standalone extension, calls Vercel API |

---

## 3. Tech Stack

| Layer | Technology | Reason |
|---|---|---|
| Framework | Next.js 16 (App Router) | Full-stack: UI + API in one repo |
| Language | TypeScript | Already configured |
| Database | MongoDB (native driver) | No ORM, flexible documents, free Atlas tier |
| UI | shadcn/ui + Tailwind CSS v4 | Already installed |
| Auth | Auth.js (NextAuth v5) | JWT sessions, credential provider |
| Mobile | Capacitor | Wraps web app, gives native push + storage |
| Extension | Manifest V3 (Vanilla JS/TS) | Chrome + Firefox compatible |
| Crypto | Web Crypto API (built-in) | Client-side password encryption |
| Push Notifications | Web Push (VAPID) | Browser + PWA notifications |
| Deployment | Vercel | Zero-config, works with Next.js |

---

## 4. Core Modules

### 4.1 Authentication
- Single-user login (username + password)
- JWT-based sessions via Auth.js
- Protected API routes using middleware
- No registration flow needed — seeded on deploy

### 4.2 Task Tracker
The core feature. Tasks are the atomic unit of Reliva.

**Fields per task:**
- Title, description
- Category (linked to a context/job)
- Priority: low / medium / high / urgent
- Status: todo / in_progress / done / archived
- Deadline (optional)
- Reminder date (optional)
- Recurrence (optional: daily, weekly, monthly)
- Tags (free-form)
- Attachments / notes (optional)

**Categories / Contexts:**
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
- Master password is **never sent to the server**
- A `derivedKey` is generated client-side from master password using PBKDF2 (Web Crypto API)
- Each credential is encrypted client-side with AES-GCM before storage
- MongoDB stores only ciphertext + IV + salt — server cannot decrypt
- Decryption happens entirely in the browser/extension

**Fields per credential:**
- Website URL
- Username / email
- Encrypted password
- Notes (optional, also encrypted)
- Icon / favicon
- Tags
- Last modified date

**Browser Extension behavior:**
- On page load, detects the current domain
- Offers to autofill matching credentials
- Lets you save new credentials directly from the extension popup
- Communicates with the Vercel API using a stored auth token

### 4.5 Notifications & Reminders
- Web Push notifications (VAPID) for browser
- Capacitor Push Plugin for mobile native notifications
- Notification types:
  - Task deadline approaching (1 day before, day-of)
  - Payroll date tomorrow
  - Upcoming travel (3 days before)
  - Overdue tasks daily digest

### 4.6 Dashboard (Home)
Unified view that shows:
- Tasks due today and this week, grouped by context
- Upcoming calendar events (next 7 days)
- Quick-add task input
- Quick credential lookup
- Overdue items highlighted

---

## 5. Database Design (MongoDB)

No ORM. All queries use the native `mongodb` driver directly.

### Collections

#### `users`
```json
{
  "_id": ObjectId,
  "username": "string",
  "passwordHash": "string",
  "pushSubscriptions": ["WebPushSubscription"],
  "createdAt": "Date"
}
```

#### `contexts` (categories/jobs)
```json
{
  "_id": ObjectId,
  "name": "Job 1 — Company Name",
  "slug": "job-1",
  "color": "#hexcode",
  "icon": "string",
  "type": "work | personal | health | finance | travel | custom",
  "order": 0
}
```

#### `tasks`
```json
{
  "_id": ObjectId,
  "title": "string",
  "description": "string",
  "contextId": ObjectId,
  "priority": "low | medium | high | urgent",
  "status": "todo | in_progress | done | archived",
  "deadline": "Date | null",
  "reminderAt": "Date | null",
  "recurrence": "none | daily | weekly | monthly | null",
  "tags": ["string"],
  "notes": "string",
  "createdAt": "Date",
  "updatedAt": "Date"
}
```

#### `events`
```json
{
  "_id": ObjectId,
  "title": "string",
  "type": "payroll | vacation | deadline | appointment | custom",
  "contextId": "ObjectId | null",
  "date": "Date",
  "endDate": "Date | null",
  "recurrence": "none | monthly | annually",
  "notes": "string",
  "createdAt": "Date"
}
```

#### `credentials`
```json
{
  "_id": ObjectId,
  "site": "string",
  "siteUrl": "string",
  "username": "string",
  "encryptedPassword": "string (base64)",
  "iv": "string (base64)",
  "salt": "string (base64)",
  "encryptedNotes": "string | null",
  "tags": ["string"],
  "lastModified": "Date",
  "createdAt": "Date"
}
```

#### `notifications`
```json
{
  "_id": ObjectId,
  "type": "task_due | event_reminder | overdue_digest",
  "refId": "ObjectId",
  "refType": "task | event",
  "message": "string",
  "sentAt": "Date",
  "read": false
}
```

---

## 6. API Route Structure (Next.js Route Handlers)

All routes are under `/app/api/`:

```
/api/auth/[...nextauth]     — Auth.js endpoints
/api/tasks                  — GET (list), POST (create)
/api/tasks/[id]             — GET, PATCH, DELETE
/api/contexts               — GET, POST
/api/contexts/[id]          — PATCH, DELETE
/api/events                 — GET, POST
/api/events/[id]            — PATCH, DELETE
/api/credentials            — GET (list, no decryption), POST
/api/credentials/[id]       — GET, PATCH, DELETE
/api/notifications          — GET, POST
/api/notifications/[id]     — PATCH (mark read)
/api/push/subscribe         — POST (save push subscription)
/api/push/send              — POST (internal, trigger push)
```

---

## 7. Frontend Page Structure (Next.js App Router)

```
app/
  layout.tsx                — Root layout (auth guard, nav)
  page.tsx                  — Dashboard / Home
  (auth)/
    login/page.tsx          — Login screen
  tasks/
    page.tsx                — Task list (filterable by context)
    [id]/page.tsx           — Task detail / edit
    new/page.tsx            — Create task
  calendar/
    page.tsx                — Calendar view
  vault/
    page.tsx                — Password manager list
    new/page.tsx            — Add credential
    [id]/page.tsx           — View/edit credential
  settings/
    page.tsx                — User settings, notification prefs
```

---

## 8. Browser Extension Structure

Lives in a separate folder: `extension/`

```
extension/
  manifest.json             — Manifest V3 (Chrome + Firefox)
  popup/
    index.html              — Extension popup UI
    popup.ts                — Logic: autofill, credential lookup
  background/
    service-worker.ts       — Background tasks, token refresh
  content/
    autofill.ts             — DOM injection for autofill
  icons/
    icon-16.png
    icon-48.png
    icon-128.png
```

The extension:
- Stores auth token in `chrome.storage.local`
- On popup open: fetches credentials matching current tab URL from API
- Injects autofill on login forms via content script
- Has a mini task quick-add form

---

## 9. Mobile (Capacitor) Strategy

After the Next.js web app is stable:

1. Run `next build` and `next export` (static output)
2. Capacitor picks up the `out/` folder
3. `npx cap add android` / `npx cap add ios`
4. Use Capacitor plugins for:
   - `@capacitor/push-notifications` — native push
   - `@capacitor/local-notifications` — reminders
   - `@capacitor/haptics` — subtle feedback
   - `@capacitor/status-bar` — mobile chrome

Mobile-specific considerations:
- The app must work with the API (Vercel URL) — not offline-first yet
- Mobile layout: bottom navigation instead of sidebar
- Responsive design from day one (Tailwind breakpoints)

---

## 10. Security Considerations

| Concern | Solution |
|---|---|
| API access | All routes protected by Auth.js session middleware |
| Password storage | Client-side AES-GCM encryption, server only holds ciphertext |
| Extension token | Stored in `chrome.storage.local`, short-lived JWT |
| Transport | HTTPS only (Vercel + Atlas enforce this) |
| MongoDB injection | Parameterized queries only (native driver) |
| CORS | Next.js API routes restrict to same origin; extension whitelisted |

---

## 11. Development Roadmap

### Phase 1 — Foundation (Current Priority)
- [ ] Set up MongoDB connection (native driver, connection pooling)
- [ ] Auth.js integration (single user, credential provider)
- [ ] Environment variables structure (`.env.local`)
- [ ] Middleware for protected routes
- [ ] Database utility layer (`lib/db.ts`)

### Phase 2 — Task Tracker (Core)
- [ ] Contexts API + UI (create work/personal/etc.)
- [ ] Tasks CRUD API
- [ ] Task list UI (Dashboard + Tasks page)
- [ ] Task detail / edit page
- [ ] Filtering, sorting, search

### Phase 3 — Calendar & Events
- [ ] Events CRUD API
- [ ] Calendar UI (month/week view)
- [ ] Tasks with deadlines appear on calendar
- [ ] Recurring events (payroll, etc.)

### Phase 4 — Password Vault
- [ ] Credentials API
- [ ] Client-side encryption/decryption (Web Crypto)
- [ ] Vault UI (list, add, view)
- [ ] Master password unlock flow (session-based key storage)

### Phase 5 — Notifications
- [ ] Web Push setup (VAPID keys)
- [ ] Push subscription API
- [ ] Cron-like triggers (Vercel Cron Jobs)
- [ ] Notification history

### Phase 6 — Browser Extension
- [ ] Manifest V3 scaffold
- [ ] Auth flow in extension (token from API)
- [ ] Credential autofill (content script)
- [ ] Quick task-add from popup
- [ ] Package for Chrome Web Store + Firefox Add-ons

### Phase 7 — Mobile
- [ ] Ensure responsive design is complete
- [ ] Capacitor setup
- [ ] Android build
- [ ] Native push notifications
- [ ] APK for personal install

---

## 12. Environment Variables

```env
# MongoDB
MONGODB_URI=mongodb+srv://...

# Auth
AUTH_SECRET=...
AUTH_URL=https://reliva.vercel.app

# Push Notifications (VAPID)
VAPID_PUBLIC_KEY=...
VAPID_PRIVATE_KEY=...
VAPID_SUBJECT=mailto:you@email.com

# Internal
CRON_SECRET=...  # Vercel cron auth header
```

---

## 13. Folder Structure (Target)

```
reliva/
  app/
    api/                    — API route handlers
    (auth)/                 — Auth pages (not in nav)
    (app)/                  — Main app pages (auth-guarded)
    layout.tsx
    page.tsx
  components/
    ui/                     — shadcn/ui primitives
    tasks/                  — Task-specific components
    vault/                  — Password vault components
    calendar/               — Calendar components
    shared/                 — Nav, layout wrappers, etc.
  lib/
    db.ts                   — MongoDB client (singleton)
    auth.ts                 — Auth.js config
    crypto.ts               — Client-side encryption utils
    push.ts                 — Web Push utilities
    validations.ts          — Zod schemas
  hooks/                    — React hooks
  extension/                — Browser extension (separate build)
  public/
  .env.local
  capacitor.config.ts       — Added in Phase 7
```

---

*This document is the single source of truth for Reliva's architecture and roadmap. Update it as the project evolves.*
