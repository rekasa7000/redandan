# Roadmap

Seven development phases, each building on the last.
Each phase produces a working, usable increment of the app.

---

## Phase 1 — Foundation ✅ Complete

**Goal:** The app boots, connects to MongoDB, requires two-factor login to access anything,
and exposes a standalone API that all three clients can authenticate against.

Shipped differently than originally planned here: the API was built as a standalone **Go + Gin**
server (`apps/server`) instead of Next.js Route Handlers, with a custom JWT issuer instead of
Auth.js. `apps/web` ended up as a pure frontend with zero database access. See ADR-011 in
`technical-decisions.md` for why, and `docs/architecture.md` / `docs/api.md` for what actually
exists. The original checklist below (`lib/db.ts`, `app/api/**`, Auth.js) was never built and should
not be — it's kept here struck through for history.

**Actually delivered:**
- [x] `apps/server/internal/db/db.go` — MongoDB singleton (Go, `sync.Once`)
- [x] `apps/server/internal/models/models.go` — Go document structs for all collections
- [x] `apps/server/internal/middleware/auth.go` — JWT validation (`RequireAuth`, `RequirePending`, `RequireCron`)
- [x] `apps/server/internal/handlers/auth.go` — Login, TOTP setup/confirm/validate, backup codes, change/forgot password, logout
- [x] `apps/server/internal/routes/routes.go` — Full route registration, `/api/v1/**`
- [x] `apps/web/middleware.ts` — Edge auth guard (checks token presence) for all `(app)` routes
- [x] `apps/web/lib/session.ts` — `reliva_token` cookie read/write/clear
- [x] `apps/web/app/(auth)/login/page.tsx` — Two-step login UI (password → TOTP/backup code)
- [x] `apps/web/app/(auth)/forgot-password/page.tsx`
- [x] `apps/web/app/(app)/layout.tsx` — Authenticated layout shell
- [x] `apps/web/app/(app)/settings/page.tsx` — TOTP setup UI (QR + backup codes), change password, logout
- [x] `apps/server/cmd/seed/main.go` — Seeds a user + default contexts; per-email, supports multiple accounts (ADR-012)
- [x] `.env.example` × 3 (`apps/server`, `apps/web`, `infra`) — documents all required env vars
- [x] `infra/docker-compose.yml` — MongoDB + server + one-shot seed service for local dev
- [x] `openapi.yaml` + `bun run gen:types` → `apps/web/lib/types.gen.ts`

~~- [ ] `lib/db.ts` — MongoDB singleton with connection pooling~~ (Go, not Next.js — see above)
~~- [ ] `lib/auth.ts` — Auth.js config + `validateCaller()` helper~~ (never built — custom Go JWT instead)
~~- [ ] `lib/totp.ts` — TOTP helpers wrapping `otplib`~~ (Go's `pquerna/otp`, server-side only)
~~- [ ] `app/api/auth/**/route.ts`~~ (no Next.js API routes exist or will exist)
~~- [ ] `vercel.json` — cron placeholder~~ (no cron on Vercel — the API isn't hosted there)

**Done when:** ✅ You can open the app, complete the two-step login (password + TOTP), see a
dashboard placeholder, and receive a `401` from the Go API on any protected endpoint without a
valid token. The extension (once built, Phase 6) will obtain a bearer token the same way the web
frontend does — posting credentials + TOTP to the same API.

---

## Phase 2 — Task Tracker ✅ Complete

**Goal:** Full task management across multiple contexts. All data flows through the standalone API.

**Backend:**
- [x] `apps/server/internal/handlers/contexts.go` — Contexts CRUD (`/api/v1/contexts/**`)
- [x] `apps/server/internal/handlers/tasks.go` — Tasks CRUD with status/context filtering (`/api/v1/tasks/**`); new tasks are created with status `"todo"` (fixed from a stray `"pending"` default found during Phase 2 work)
- [x] List handlers (`tasks`, `contexts`, `events`, `credentials`, `notifications`) initialize slices as `[]T{}` instead of `var x []T`, so empty lists serialize as JSON `[]` — Go's zero-value `nil` slice otherwise serializes as `null`, which crashed the frontend on a fresh account with zero tasks

**Frontend:**
- [x] `apps/web/lib/api.ts` — typed fetch wrapper around the Go API (using `lib/types.gen.ts`), coerces the `null`-vs-`[]` list response
- [x] `apps/web/lib/validations.ts` — Zod schema for the task form
- [x] `apps/web/lib/utils.ts` — timezone-safe date helpers (`toDateInputValue`, `fromDateInputValue`, `todayDateString`, `addDaysToDateString`); due dates are compared as `"YYYY-MM-DD"` strings, never `Date` objects, to avoid local-timezone day-boundary drift
- [x] `apps/web/app/(app)/page.tsx` — Dashboard: Overdue / Due today / Upcoming (7-day) sections
- [x] `apps/web/app/(app)/tasks/page.tsx` — Task list with status/context filters (reflected in the URL)
- [x] `apps/web/app/(app)/tasks/new/page.tsx` — Create task form
- [x] `apps/web/app/(app)/tasks/[id]/page.tsx` — Task detail / edit / delete
- [x] `apps/web/components/tasks/task-card.tsx` — quick done-toggle, priority/status badges, context color dot
- [x] `apps/web/components/tasks/task-form.tsx` — shared create/edit form
- [x] `apps/web/components/tasks/task-filters.tsx`
- [x] `apps/web/components/shared/sidebar.tsx` — nav + context list (desktop)
- [x] `apps/web/components/shared/bottom-nav.tsx` — mobile bottom nav

**Also fixed along the way:** `openapi.yaml`'s schemas (`Context`, `Task`, `Event`, `Credential`, `Notification`) didn't match the Go handlers' actual JSON output at all — camelCase vs snake_case, `_id` vs `id`, wrong field names (`deadline` vs `due_date`), and an entirely invented `Notification` shape. Fixed and regenerated `lib/types.gen.ts` before building anything against it — see `docs/api.md` for the corrected contract.

**Done when:** ✅ You can create a task under "Work", set a due date and priority, see it on the dashboard grouped by Overdue/Due today/Upcoming, toggle it done from the list, edit it, and delete it — verified end-to-end in a real browser session, not just typecheck/lint.

---

## Phase 3 — Calendar & Events

**Goal:** A unified timeline showing tasks with deadlines, payroll dates, and personal events.

**Backend — already done:**
- [x] `apps/server/internal/handlers/events.go` — Events CRUD (`/api/v1/events/**`)

**Frontend — remaining work:**
- [ ] `apps/web/app/(app)/calendar/page.tsx` — Calendar view (month + agenda)
- [ ] `apps/web/components/calendar/calendar-view.tsx`
- [ ] `apps/web/components/calendar/event-item.tsx`
- [ ] Tasks with deadlines rendered on the calendar alongside events
- [ ] Recurring event support (payroll monthly, vacation annually) — the `Event.Recurrence` field
      exists in the Go model but isn't acted on by any handler yet; recurrence expansion will need
      to be built, either in the Go handler or client-side

**Done when:** You can add "Payroll — Work" as a monthly recurring event, add a "Cebu trip" multi-day event, and see both alongside task deadlines on the calendar.

---

## Phase 4 — Password Vault

**Goal:** Encrypted credential storage with a master-password-gated vault UI.

**Backend — already done:**
- [x] `apps/server/internal/handlers/credentials.go` — Credentials CRUD, stores/returns ciphertext only (`/api/v1/credentials/**`)

**Frontend — remaining work:**
- [ ] `apps/web/lib/crypto.ts` — PBKDF2 key derivation + AES-GCM encrypt/decrypt utilities
- [ ] `apps/web/hooks/use-master-key.ts` — In-memory key management, lock/unlock
- [ ] `apps/web/app/(app)/vault/page.tsx` — Credential list
- [ ] `apps/web/app/(app)/vault/new/page.tsx` — Add credential (encrypts before POST)
- [ ] `apps/web/app/(app)/vault/[id]/page.tsx` — View credential (decrypt on demand)
- [ ] `apps/web/components/vault/master-password-gate.tsx` — Unlock prompt
- [ ] `apps/web/components/vault/credential-card.tsx` — Copy-to-clipboard
- [ ] `apps/web/components/vault/password-generator.tsx` — Random strong password generator

**Done when:** You can add a GitHub credential, lock the vault, re-enter the master password, and see the decrypted password. MongoDB shows only ciphertext — no plaintext visible anywhere.

---

## Phase 5 — Notifications

**Goal:** Proactive alerts for deadlines, payroll, and overdue tasks.

**Backend — already done:**
- [x] `apps/server/internal/handlers/notifications.go` — `GET /api/v1/notifications`, `PATCH /api/v1/notifications/:id`, `POST /api/v1/push/subscribe`
- [x] `apps/server/internal/handlers/cron.go` — `POST /api/v1/cron/notify`, creates a notification per task due today across all users

**Backend — remaining work:**
- [ ] VAPID key generation (one-time: `npx web-push generate-vapid-keys`) and wiring `VAPID_*` env vars through to an actual send
- [ ] Web Push send logic — `push_subscriptions` are stored, but nothing sends to them yet; `cron.go` only creates in-app `notifications` documents today
- [ ] An external scheduler configured to call `POST /api/v1/cron/notify` daily (Railway cron / a scheduled GitHub Action — **not** Vercel Cron, since the API isn't hosted on Vercel)

**Frontend — remaining work:**
- [ ] Service worker registration + push subscription flow in `apps/web`
- [ ] `apps/web/components/shared/notification-bell.tsx` — Unread badge in nav

**Done when:** At 8am you receive a push notification listing tasks due today. The notification bell in the app shows the history.

---

## Phase 6 — Browser Extension

**Goal:** Chrome and Firefox extension for password autofill and quick task creation.

**Deliverables:**
- [ ] `extension/manifest.json` — Manifest V3
- [ ] `extension/popup/index.html` + `popup.ts` — Mini UI
- [ ] `extension/background/service-worker.ts` — Token management, badge count
- [ ] `extension/content/autofill.ts` — Login form detection + fill
- [ ] Two-step auth flow in extension (password + TOTP → bearer token)
- [ ] Token stored in `chrome.storage.local`; refreshed by service worker
- [ ] Domain-matching credential lookup
- [ ] Quick task-add form in popup
- [ ] Build script (`esbuild`) in `package.json`

**Done when:** You visit GitHub's login page, click the Reliva extension icon, see your GitHub credentials, click to autofill, and the form is filled.

---

## Phase 7 — Mobile (Android)

**Goal:** Install Reliva on your Android phone as a native app with push notifications.

**Deliverables:**
- [ ] `next.config.ts` updated with `output: 'export'`
- [ ] All data-fetching pages converted to client-side fetching
- [ ] `capacitor.config.ts` — pointing to Vercel URL
- [ ] Android project generated (`bun cap add android`)
- [ ] `@capacitor/push-notifications` configured
- [ ] `@capacitor/local-notifications` for scheduled reminders
- [ ] `@capacitor/status-bar` and `@capacitor/haptics`
- [ ] APK built and sideloaded on personal Android device
- [ ] Mobile-specific UI tweaks (bottom nav, larger tap targets)

**Done when:** Reliva is installed on your phone. You get a native push at 8am. You can create tasks and check the calendar from the app.

---

## Future Ideas (Post-MVP)

- **Offline support** — Cache in IndexedDB, sync when back online
- **Rich text notes** — Markdown or WYSIWYG per-task notes
- **File attachments** — Stored in Vercel Blob or S3
- **Quick capture** — Share sheet on Android to create tasks from any app
- **Home screen widget** — Surface today's tasks without opening the app
- **Export / backup** — JSON export of all data for personal archive
- **iOS build** — Requires macOS + Xcode
- **WebAuthn / Passkeys** — Hardware key as a TOTP replacement (higher security ceiling)
- **Self-serve signup** — The data model is already multi-user (ADR-012); what's still missing to
  become a real product for others is signup, per-account roles, rate limiting, and a unique index
  on `users.email`

---

## Phase Status Tracker

| Phase | Status |
|---|---|
| Phase 1 — Foundation + 2FA | ✅ Complete |
| Phase 2 — Task Tracker | ✅ Complete |
| Phase 3 — Calendar & Events | Backend done; frontend not started |
| Phase 4 — Password Vault | Backend done; frontend not started |
| Phase 5 — Notifications | Backend partially done (no push send yet); frontend not started |
| Phase 6 — Browser Extension | Not started |
| Phase 7 — Mobile | Not started |

Update this table as phases are completed.
