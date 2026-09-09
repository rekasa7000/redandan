# Platforms

Reliva runs on three platforms from a single monorepo. Each platform consumes the same standalone
Go API (`apps/server`) — none of them is hosted on the same platform as the API itself.

---

## Platform 1 — Web (Primary)

**Stack:** Next.js 16 (pure frontend) on Vercel
**API:** Go server, deployed separately (Railway/Fly.io) — see `docs/architecture.md`
**URL:** `https://reliva.vercel.app` (or custom domain)
**Status:** Phase 1 done (auth); Phase 2+ in progress

The web app is the primary development target and the source of truth for UI and features.
All other platforms derive from it. It has no API routes and no database access — every page that
needs data is a Client Component calling the Go API with `fetch` and an `Authorization: Bearer`
header.

### Responsive Design Strategy

The UI must be mobile-first from day one, even before Capacitor is added:

| Breakpoint | Layout |
|---|---|
| `< 768px` (mobile) | Bottom navigation bar, single-column layout |
| `768px–1024px` (tablet) | Sidebar collapsed, two-column |
| `> 1024px` (desktop) | Sidebar expanded, full multi-column |

Use Tailwind's responsive prefixes (`md:`, `lg:`) to implement this progressively.
Avoid fixed pixel widths. Use `max-w-*` containers.

### PWA Capability

Before Capacitor (Phase 7), the web app can be installed as a Progressive Web App on Android via
Chrome:
- Add a `public/manifest.json` with app name, icons, `display: standalone`
- Register a service worker for offline caching (optional, basic shell caching)
- This gives a "homescreen icon" experience without building an APK

### Vercel Deployment

Deploys automatically on push to `main`. Project Root Directory is set to `apps/web` in Vercel's
project settings. There is no `vercel.json` cron config for notifications — that would only matter
if the API itself ran on Vercel, and it doesn't (see `docs/architecture.md`).

---

## Platform 2 — Browser Extension

**Targets:** Chrome (Manifest V3), Firefox (Manifest V3)
**Status:** Phase 6 — not started, `apps/extension/` is a placeholder
**Location:** `apps/extension/` in the repo

### What the Extension Will Do

1. **Password Autofill** — Detects login forms on any page, fetches matching credentials from the
   Go API, fills username and password with one click
2. **Quick Task Add** — Create a task from any tab without opening the full app
3. **Notification Badge** — Shows count of overdue tasks / unread notifications on the extension icon

### Planned Architecture

```
apps/extension/
  manifest.json           ← Manifest V3 config
  popup/
    index.html            ← Popup UI (shown when clicking the extension icon)
    popup.ts              ← Fetches credentials, renders UI, handles quick-add
  background/
    service-worker.ts     ← Token refresh, badge update
  content/
    autofill.ts            ← Injected into every page, detects login forms
```

### Planned Auth Flow in Extension

Same two-step flow as every other client — no special extension-only endpoint:
1. Popup collects email + password → `POST /api/v1/auth/login` on the Go server
2. Popup collects TOTP code → `POST /api/v1/auth/totp/validate`
3. Resulting access JWT is stored in `chrome.storage.local`
4. Every subsequent API call sends `Authorization: Bearer <token>`
5. Background service worker prompts re-login before the 30-day token expires (there's no silent
   refresh endpoint — a new login is required)

### Planned Autofill Flow

1. Content script runs on every page load
2. Detects `<input type="password">` — signals a login form
3. Fetches `GET /api/v1/credentials` with the stored token, matches by domain client-side (the API
   has no `?site=` filter today — see `docs/api.md`)
4. If matches found, shows a small Reliva icon inside the input field
5. User clicks icon → popup shows matched credentials, decrypted client-side → user clicks to fill

### Cross-Browser Notes

Firefox and Chrome both support Manifest V3 but with some differences:
- Firefox uses the `browser.*` namespace (polyfilled with `webextension-polyfill`)
- `service_worker` in `background` is Chrome; Firefox uses `scripts` instead
- `manifest.json` can include Firefox-specific keys under `browser_specific_settings`

Build with `esbuild` for fast, simple bundling. No webpack needed.

---

## Platform 3 — Mobile (Android)

**Stack:** Capacitor wrapping a static Next.js export
**Target:** Android (personal sideload)
**Status:** Phase 7 — not started

### How Capacitor Will Work

```
Next.js build (static export)
  → copied into Capacitor's www/ folder
  → Capacitor wraps it in an Android Activity
  → Native APIs available via @capacitor/* plugins
```

### Planned Setup Steps (Phase 7)

1. Set `output: 'export'` in `apps/web/next.config.ts`
2. `next build` → generates `out/`
3. `bun add @capacitor/core @capacitor/cli`
4. `bun cap init reliva com.personal.reliva`
5. Set `webDir: 'out'` in `capacitor.config.ts`, and `server.url` to the deployed Go API's frontend
   origin (i.e. still the Vercel URL — the app loads the web UI, which then calls the Go API)
6. `bun cap add android` → `bun cap sync` → `bun cap open android`
7. Build APK → sideload

### Planned Native Features

| Plugin | Purpose |
|---|---|
| `@capacitor/push-notifications` | Receive push notifications when app is closed |
| `@capacitor/local-notifications` | Schedule local reminders (e.g., deadline at 9am) |
| `@capacitor/haptics` | Subtle vibration feedback on task completion |
| `@capacitor/status-bar` | Match status bar to app theme |
| `@capacitor/app` | Handle back button, app state |

### API Connectivity

The mobile app calls the same Go API as the web app, using the same bearer-token auth — not a
session cookie (Capacitor's WebView doesn't share cookies with the API's origin anyway). The token
is stored in Capacitor's secure storage, not a cookie.

```ts
const config: CapacitorConfig = {
  appId: 'com.personal.reliva',
  appName: 'Reliva',
  webDir: 'out',
  server: { url: 'https://reliva.vercel.app', cleartext: false },
};
```

This means the app requires an internet connection. Offline support is not a Phase 7 goal.

### Static Export Constraints

When Next.js uses `output: 'export'`, all data-fetching must already be client-side — which is
already true today, since `apps/web` has no server components doing data fetching and no API routes
to lose. Phase 7 should be a smaller lift than originally scoped for exactly this reason.

---

## Platform Comparison

| Feature | Web | Extension | Mobile |
|---|---|---|---|
| Task management | Full | Quick-add only | Full |
| Password vault | Full | Autofill + add | View only |
| Calendar | Full | No | Full |
| Push notifications | Via Web Push | No | Native |
| Offline support | No | No | No |
| Auth method | Bearer JWT (cookie-persisted) | Bearer JWT (`chrome.storage.local`) | Bearer JWT (Capacitor secure storage) |
