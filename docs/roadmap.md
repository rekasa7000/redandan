# Roadmap

Seven development phases, each building on the last.
Each phase produces a working, usable increment of the app.

---

## Phase 1 — Foundation

**Goal:** The app boots, connects to MongoDB, and requires login to access anything.

**Deliverables:**
- [ ] `lib/db.ts` — MongoDB singleton with connection pooling
- [ ] `lib/types.ts` — TypeScript interfaces for all document types
- [ ] `lib/auth.ts` — Auth.js config with credentials provider
- [ ] `middleware.ts` — Auth guard for all `(app)` routes
- [ ] `app/(auth)/login/page.tsx` — Login page
- [ ] `app/(app)/layout.tsx` — Authenticated layout shell (nav, sidebar stubs)
- [ ] `scripts/seed.ts` — Seeds admin user and default contexts
- [ ] `.env.example` — Documents all required environment variables
- [ ] `vercel.json` — Base config (cron placeholder)

**Done when:** You can open the app, see the login page, log in, and see a blank dashboard. Visiting any protected route while logged out redirects to login.

---

## Phase 2 — Task Tracker

**Goal:** Full task management across multiple contexts.

**Deliverables:**
- [ ] `app/api/contexts/**` — Contexts CRUD
- [ ] `app/api/tasks/**` — Tasks CRUD with filtering
- [ ] `app/(app)/page.tsx` — Dashboard: today's tasks + upcoming
- [ ] `app/(app)/tasks/page.tsx` — Task list with context/status/priority filters
- [ ] `app/(app)/tasks/new/page.tsx` — Create task form
- [ ] `app/(app)/tasks/[id]/page.tsx` — Task detail / edit
- [ ] `components/tasks/task-card.tsx` — Task card component
- [ ] `components/tasks/task-form.tsx` — Shared form (create + edit)
- [ ] `components/tasks/task-filters.tsx` — Filter bar
- [ ] `components/shared/sidebar.tsx` — Sidebar with context list
- [ ] `components/shared/bottom-nav.tsx` — Mobile bottom nav
- [ ] `lib/validations.ts` — Zod schemas for tasks and contexts

**Done when:** You can create a task under "Job 1", mark it in-progress, set a deadline, and see it on the dashboard.

---

## Phase 3 — Calendar & Events

**Goal:** A unified timeline showing tasks with deadlines, payroll dates, and personal events.

**Deliverables:**
- [ ] `app/api/events/**` — Events CRUD
- [ ] `app/(app)/calendar/page.tsx` — Calendar view (month + agenda)
- [ ] `components/calendar/calendar-view.tsx` — Month grid
- [ ] `components/calendar/event-item.tsx` — Event display
- [ ] Tasks with deadlines rendered on the calendar alongside events
- [ ] Recurring event support (payroll monthly, vacation annually)

**Done when:** You can add "Payroll — Job 1" as a monthly recurring event on the 15th, add a "Cebu trip" multi-day event, and see both alongside your task deadlines on the calendar.

---

## Phase 4 — Password Vault

**Goal:** Encrypted credential storage with a master-password-gated vault UI.

**Deliverables:**
- [ ] `lib/crypto.ts` — PBKDF2 key derivation + AES-GCM encrypt/decrypt utilities
- [ ] `hooks/use-master-key.ts` — In-memory key management, lock/unlock
- [ ] `app/api/credentials/**` — Credentials CRUD (store/return ciphertext only)
- [ ] `app/(app)/vault/page.tsx` — Credential list (shows site + username, password hidden)
- [ ] `app/(app)/vault/new/page.tsx` — Add credential (encrypts before POST)
- [ ] `app/(app)/vault/[id]/page.tsx` — View credential (decrypt on demand)
- [ ] `components/vault/master-password-gate.tsx` — Unlock prompt shown before vault access
- [ ] `components/vault/credential-card.tsx` — Card with copy-to-clipboard
- [ ] `components/vault/password-generator.tsx` — Random strong password generator

**Done when:** You can add a credential for GitHub, lock the vault, re-enter your master password, and see the decrypted password. The MongoDB Atlas dashboard shows only ciphertext for all credential documents.

---

## Phase 5 — Notifications

**Goal:** Proactive alerts for deadlines, payroll, and overdue tasks across browser and (later) mobile.

**Deliverables:**
- [ ] VAPID key generation (one-time: `npx web-push generate-vapid-keys`)
- [ ] `lib/push.ts` — Web Push send utility
- [ ] `app/api/push/subscribe/route.ts` — Save push subscription
- [ ] `app/api/push/send/route.ts` — Internal trigger
- [ ] `app/api/cron/notify/route.ts` — Cron handler: query due/overdue + send push
- [ ] `vercel.json` — Cron schedule (daily 8am)
- [ ] `app/api/notifications/**` — Notification history CRUD
- [ ] `components/shared/notification-bell.tsx` — Unread badge in nav
- [ ] Service worker registration in the web app (for push subscription)

**Done when:** At 8am, you receive a push notification in your browser listing tasks due today and any payroll event tomorrow. The notification bell in the app shows the history.

---

## Phase 6 — Browser Extension

**Goal:** Chrome and Firefox extension for password autofill and quick task creation.

**Deliverables:**
- [ ] `extension/manifest.json` — Manifest V3 for Chrome + Firefox
- [ ] `extension/popup/index.html` + `popup.ts` — Mini UI
- [ ] `extension/background/service-worker.ts` — Token management, badge count
- [ ] `extension/content/autofill.ts` — Login form detection + fill
- [ ] Extension auth flow (token exchange with the Vercel API)
- [ ] Domain-matching credential lookup
- [ ] Quick task-add form in popup
- [ ] Build script (`esbuild` or similar) in `package.json`
- [ ] Load unpacked in Chrome for personal use (no store submission required)

**Done when:** You visit GitHub's login page, click the Redandan extension icon, see your GitHub credentials listed, click to autofill, and the form is filled automatically.

---

## Phase 7 — Mobile (Android)

**Goal:** Install Redandan on your Android phone as a native app with push notifications.

**Deliverables:**
- [ ] `next.config.ts` updated with `output: 'export'`
- [ ] All data-fetching pages converted to client-side fetching (API calls on mount)
- [ ] `capacitor.config.ts` — App config pointing to Vercel URL
- [ ] Android project generated (`bun cap add android`)
- [ ] `@capacitor/push-notifications` configured
- [ ] `@capacitor/local-notifications` for scheduled reminders
- [ ] `@capacitor/status-bar` and `@capacitor/haptics` for polish
- [ ] APK built and sideloaded on personal Android device
- [ ] Mobile-specific UI tweaks (bottom nav, larger tap targets)

**Done when:** Redandan is installed on your phone as an app. You get a native push notification at 8am. You can create tasks and check your calendar from your phone.

---

## Future Ideas (Post-MVP)

These are not on the current roadmap but worth noting for later:

- **Offline support** — Cache tasks/events in IndexedDB, sync when back online
- **Rich text notes** — Per-task notes with markdown or WYSIWYG editor
- **File attachments** — Attach files to tasks (stored in Vercel Blob or S3)
- **Quick capture** — Share sheet on mobile to create tasks from any app
- **Apple Watch / widget** — Surface today's tasks on a watch face or home screen widget
- **Multi-device sync indicator** — Show last-sync timestamp
- **Export / backup** — JSON export of all data for personal archive
- **iOS build** — Requires macOS + Xcode
- **Multi-user** — If Redandan ever becomes a product for others

---

## Phase Status Tracker

| Phase | Status |
|---|---|
| Phase 1 — Foundation | Not started |
| Phase 2 — Task Tracker | Not started |
| Phase 3 — Calendar & Events | Not started |
| Phase 4 — Password Vault | Not started |
| Phase 5 — Notifications | Not started |
| Phase 6 — Browser Extension | Not started |
| Phase 7 — Mobile | Not started |

Update this table as phases are completed.
