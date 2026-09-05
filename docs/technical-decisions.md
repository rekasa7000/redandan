# Technical Decisions

Architecture Decision Records (ADRs) for Redandan.
Each decision documents the context, the choice made, and why.

---

## ADR-001: No ORM — Use MongoDB Native Driver

**Status:** Decided

**Context:**
The project uses MongoDB. Common choices would be Mongoose (ODM) or Prisma (with MongoDB connector). Both add an abstraction layer on top of the native driver.

**Decision:**
Use the official `mongodb` npm package directly. Write all queries in TypeScript against the native driver API.

**Reasoning:**
- This is a personal project — simplicity matters more than convention
- Mongoose's schema system duplicates what TypeScript interfaces already provide
- Prisma's MongoDB connector is not as mature as its SQL connectors
- Direct queries give full access to MongoDB features (aggregation pipelines, partial indexes, etc.) without ORM limitations
- Less magic: what you write is exactly what gets sent to MongoDB

**Consequences:**
- Document shapes are defined as TypeScript interfaces in `lib/types.ts`
- Validation happens via Zod before data reaches the database
- Queries are more verbose but fully transparent

---

## ADR-002: Client-Side Encryption for Password Vault

**Status:** Decided

**Context:**
The password vault stores credentials. These could be encrypted server-side (the server holds the key) or client-side (the browser holds the key).

**Decision:**
Encrypt and decrypt credentials entirely on the client using the Web Crypto API. The server stores only ciphertext, IV, and salt. The master password never leaves the browser.

**Reasoning:**
- Server-side encryption means the server can decrypt your passwords — this is a single point of compromise
- If the MongoDB database is leaked, attacker gets only useless ciphertext
- If Vercel is compromised, same — no keys, no plaintext
- Web Crypto API is available natively in all modern browsers and Capacitor WebViews
- No dependency on a third-party crypto library

**Consequences:**
- If you forget the master password, vault data is permanently unrecoverable (by design)
- Master key is kept in memory only (sessionStorage key handle) and must be re-entered after browser restart
- The browser extension must also implement the same decryption logic
- Server cannot search or filter by credential content (only by unencrypted metadata like `site` and `username`)

---

## ADR-003: Auth.js for Authentication

**Status:** Decided

**Context:**
The app needs authentication. Options: Auth.js (NextAuth v5), custom JWT implementation, Clerk, Supabase Auth.

**Decision:**
Use Auth.js (NextAuth v5) with a credentials provider.

**Reasoning:**
- Clerk and Supabase Auth are third-party services — unnecessary dependency for a personal app
- A fully custom JWT implementation is error-prone (token rotation, CSRF, secure cookies)
- Auth.js integrates natively with Next.js App Router and middleware
- The credentials provider supports username/password, which is all that's needed
- JWT sessions (not database sessions) work well with Vercel's stateless serverless functions

**Consequences:**
- User is seeded into the database at setup — no registration flow
- Session is available server-side via `auth()` and client-side via `useSession()`
- Password reset flow needs to be built manually if ever needed (currently out of scope)

---

## ADR-004: Single Repository for Web and Extension

**Status:** Decided

**Context:**
The browser extension is a separate artifact from the Next.js web app. It could live in a separate repository.

**Decision:**
Keep the extension in `extension/` within the main repo. It is a separate build but shares the same git history and types.

**Reasoning:**
- Shared TypeScript types between the web app and extension (e.g., credential shapes, API response types)
- One place to see all platform changes together
- Simpler for a solo developer — no need to coordinate between two repos
- The extension is relatively small and closely coupled to the API

**Consequences:**
- The extension has its own `build` script and `manifest.json`
- It does not use Next.js — it's built with a simple bundler (e.g., `esbuild`)
- Both builds are triggered from the root `package.json` if needed

---

## ADR-005: Capacitor for Mobile

**Status:** Decided (Phase 7)

**Context:**
The app needs to run on mobile. Options: React Native (rewrite), Expo, Capacitor, or PWA-only.

**Decision:**
Use Capacitor to wrap the existing Next.js web app into a native Android/iOS shell.

**Reasoning:**
- React Native requires rewriting all UI — high cost for no additional benefit for a personal app
- Capacitor wraps the existing web UI, so all design work is shared
- Capacitor provides native plugins for push notifications, haptics, and status bar — everything needed
- The app is already designed to be responsive (Tailwind mobile-first)
- PWA-only is not viable for personal Android install without sideloading complexity

**Consequences:**
- Next.js must produce a static export (`output: 'export'` in `next.config.ts`) for Capacitor to consume
- API calls go to the Vercel URL (not local) — internet connection required
- Android build requires Android Studio and Java SDK on the dev machine
- iOS build requires macOS and Xcode (not a current priority)

---

## ADR-006: Vercel Cron Jobs for Notifications

**Status:** Decided

**Context:**
Scheduled notifications (daily digest, deadline reminders) need a mechanism to trigger on a schedule. Options: Vercel Cron Jobs, a separate cron server, database-triggered functions, or a third-party scheduler.

**Decision:**
Use Vercel Cron Jobs configured in `vercel.json`.

**Reasoning:**
- No additional infrastructure required — Vercel already hosts the app
- Cron Jobs call a protected API route (`/api/cron/notify`) with a secret header
- Free tier includes up to 2 cron jobs running daily
- Simple, auditable, and co-located with the app

**Consequences:**
- Cron is limited to daily granularity on Vercel's free tier
- The cron route must be secured with a `CRON_SECRET` header check to prevent unauthorized triggering
- If Vercel is down at the scheduled time, the notification is missed (acceptable for personal use)

---

## ADR-007: Manifest V3 for Browser Extension

**Status:** Decided

**Context:**
Browser extensions can use Manifest V2 (legacy, being deprecated) or Manifest V3 (current standard).

**Decision:**
Build the extension targeting Manifest V3.

**Reasoning:**
- Manifest V2 is deprecated in Chrome (phased out) and will be removed
- Firefox now supports Manifest V3 (with some differences, but manageable)
- MV3 is the future-proof choice
- The extension's needs (content script autofill, background service worker, storage) are all well-supported in MV3

**Consequences:**
- Background pages become service workers in MV3 — they can be terminated by the browser
- Persistent state must use `chrome.storage` rather than in-memory variables
- The content script security model is stricter — CSP must be configured correctly
- Firefox MV3 compatibility requires testing (minor API differences)

---

## ADR-008: Route Groups for Auth-Protected Pages

**Status:** Decided

**Context:**
Some pages require authentication (all app pages), and some do not (login). Next.js App Router needs a clean way to apply the auth guard.

**Decision:**
Use route groups: `(auth)` for public auth pages, `(app)` for all protected pages. Middleware applies the auth check to all routes except those under `(auth)`.

**Reasoning:**
- Route groups (parentheses folders) are invisible in the URL
- Clean separation between authenticated and unauthenticated surfaces
- Middleware runs at the edge, before any rendering — minimal latency for auth checks
- No need for per-page auth checks — the middleware handles it globally

**Consequences:**
- All pages must be placed under the correct route group
- Middleware must maintain an updated list of public routes (or use a path-based pattern)
