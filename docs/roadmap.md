# Roadmap

Seven development phases, each building on the last.
Each phase produces a working, usable increment of the app.

---

## Phase 1 — Foundation

**Goal:** The app boots, connects to MongoDB, requires two-factor login to access anything,
and exposes a standalone API that all three clients can authenticate against.

**Deliverables:**
- [ ] `lib/db.ts` — MongoDB singleton with connection pooling
- [ ] `lib/types.ts` — TypeScript interfaces for all document types
- [ ] `lib/auth.ts` — Auth.js config + `validateCaller()` helper (cookie + bearer)
- [ ] `lib/totp.ts` — TOTP helpers wrapping `otplib` (generate secret, verify code)
- [ ] `middleware.ts` — Auth guard for all `(app)` routes
- [ ] `app/api/auth/login/route.ts` — Step 1: password validation
- [ ] `app/api/auth/totp/validate/route.ts` — Step 2: TOTP validation, issues session or JWT
- [ ] `app/api/auth/totp/setup/route.ts` — Generates secret + QR code (settings use)
- [ ] `app/api/auth/totp/confirm/route.ts` — Confirms TOTP setup
- [ ] `app/api/auth/backup-code/route.ts` — Uses a backup code in place of TOTP
- [ ] `app/api/auth/logout/route.ts` — Invalidates session
- [ ] `app/(auth)/login/page.tsx` — Two-step login UI (password → TOTP code)
- [ ] `app/(app)/layout.tsx` — Authenticated layout shell (nav, sidebar stubs)
- [ ] `app/(app)/settings/page.tsx` — TOTP setup UI (QR code + backup codes)
- [ ] `scripts/seed.ts` — Seeds admin user, hashed password, default contexts
- [ ] `.env.example` — Documents all required environment variables
- [ ] `vercel.json` — Base config (cron placeholder, CORS headers)

**Done when:** You can open the app, complete the two-step login (password + TOTP), see a blank
dashboard, and receive a `401` on any API endpoint when calling without a valid session or token.
The extension can obtain a bearer token by posting credentials + TOTP to the API.

---

## Phase 2 — Task Tracker

**Goal:** Full task management across multiple contexts. All data flows through the standalone API.

**Deliverables:**
- [ ] `app/api/contexts/**` — Contexts CRUD
- [ ] `app/api/tasks/**` — Tasks CRUD with filtering
- [ ] `app/(app)/page.tsx` — Dashboard: today's tasks + upcoming
- [ ] `app/(app)/tasks/page.tsx` — Task list with context/status/priority filters
- [ ] `app/(app)/tasks/new/page.tsx` — Create task form
- [ ] `app/(app)/tasks/[id]/page.tsx` — Task detail / edit
- [ ] `components/tasks/task-card.tsx`
- [ ] `components/tasks/task-form.tsx`
- [ ] `components/tasks/task-filters.tsx`
- [ ] `components/shared/sidebar.tsx` — Context list
- [ ] `components/shared/bottom-nav.tsx` — Mobile bottom nav
- [ ] `lib/validations.ts` — Zod schemas for tasks and contexts

**Done when:** You can create a task under "Job 1", mark it in-progress, set a deadline, and see it on the dashboard. The same task is fetchable via `GET /api/tasks` with a bearer token.

---

## Phase 3 — Calendar & Events

**Goal:** A unified timeline showing tasks with deadlines, payroll dates, and personal events.

**Deliverables:**
- [ ] `app/api/events/**` — Events CRUD
- [ ] `app/(app)/calendar/page.tsx` — Calendar view (month + agenda)
- [ ] `components/calendar/calendar-view.tsx`
- [ ] `components/calendar/event-item.tsx`
- [ ] Tasks with deadlines rendered on the calendar alongside events
- [ ] Recurring event support (payroll monthly, vacation annually)

**Done when:** You can add "Payroll — Job 1" as a monthly recurring event, add a "Cebu trip" multi-day event, and see both alongside task deadlines on the calendar.

---

## Phase 4 — Password Vault

**Goal:** Encrypted credential storage with a master-password-gated vault UI.

**Deliverables:**
- [ ] `lib/crypto.ts` — PBKDF2 key derivation + AES-GCM encrypt/decrypt utilities
- [ ] `hooks/use-master-key.ts` — In-memory key management, lock/unlock
- [ ] `app/api/credentials/**` — Credentials CRUD (store/return ciphertext only)
- [ ] `app/(app)/vault/page.tsx` — Credential list
- [ ] `app/(app)/vault/new/page.tsx` — Add credential (encrypts before POST)
- [ ] `app/(app)/vault/[id]/page.tsx` — View credential (decrypt on demand)
- [ ] `components/vault/master-password-gate.tsx` — Unlock prompt
- [ ] `components/vault/credential-card.tsx` — Copy-to-clipboard
- [ ] `components/vault/password-generator.tsx` — Random strong password generator

**Done when:** You can add a GitHub credential, lock the vault, re-enter the master password, and see the decrypted password. MongoDB shows only ciphertext — no plaintext visible anywhere in Atlas.

---

## Phase 5 — Notifications

**Goal:** Proactive alerts for deadlines, payroll, and overdue tasks.

**Deliverables:**
- [ ] VAPID key generation (one-time: `npx web-push generate-vapid-keys`)
- [ ] `lib/push.ts` — Web Push send utility
- [ ] `app/api/push/subscribe/route.ts` — Save push subscription
- [ ] `app/api/push/send/route.ts` — Internal trigger
- [ ] `app/api/cron/notify/route.ts` — Cron handler
- [ ] `vercel.json` — Cron schedule (daily 8am)
- [ ] `app/api/notifications/**` — Notification history CRUD
- [ ] `components/shared/notification-bell.tsx` — Unread badge in nav
- [ ] Service worker registration in the web app

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

**Done when:** You visit GitHub's login page, click the Redandan extension icon, see your GitHub credentials, click to autofill, and the form is filled.

---

## Phase 7 — Mobile (Android)

**Goal:** Install Redandan on your Android phone as a native app with push notifications.

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

**Done when:** Redandan is installed on your phone. You get a native push at 8am. You can create tasks and check the calendar from the app.

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
- **Multi-user** — If Redandan ever becomes a product for others

---

## Phase Status Tracker

| Phase | Status |
|---|---|
| Phase 1 — Foundation + 2FA | Not started |
| Phase 2 — Task Tracker | Not started |
| Phase 3 — Calendar & Events | Not started |
| Phase 4 — Password Vault | Not started |
| Phase 5 — Notifications | Not started |
| Phase 6 — Browser Extension | Not started |
| Phase 7 — Mobile | Not started |

Update this table as phases are completed.
