# Platforms

Redandan runs on three platforms from a single codebase.
Each platform consumes the same Vercel-hosted API.

---

## Platform 1 — Web (Primary)

**Stack:** Next.js 16 on Vercel
**URL:** `https://redandan.vercel.app` (or custom domain)
**Status:** Phase 1–5

The web app is the primary development target and the source of truth for UI and features.
All other platforms derive from it.

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

Before Capacitor (Phase 7), the web app can be installed as a Progressive Web App on Android via Chrome:

- Add a `public/manifest.json` with app name, icons, `display: standalone`
- Register a service worker for offline caching (optional, basic shell caching)
- This gives a "homescreen icon" experience without building an APK

### Vercel Deployment

Deploys automatically on push to `main`.

Configuration in `vercel.json`:
```json
{
  "crons": [
    {
      "path": "/api/cron/notify",
      "schedule": "0 8 * * *"
    }
  ]
}
```

---

## Platform 2 — Browser Extension

**Targets:** Chrome (Manifest V3), Firefox (Manifest V3)
**Status:** Phase 6
**Location:** `extension/` folder in the repo

### What the Extension Does

1. **Password Autofill** — Detects login forms on any page, fetches matching credentials from the API, fills username and password with one click
2. **Quick Task Add** — Create a task from any tab without opening the full app
3. **Notification Badge** — Shows count of overdue tasks / unread notifications on the extension icon

### Architecture

```
extension/
  manifest.json           ← Manifest V3 config
  popup/
    index.html            ← Popup UI (shown when clicking the extension icon)
    popup.ts              ← Fetches credentials, renders UI, handles quick-add
  background/
    service-worker.ts     ← Token refresh, badge update
  content/
    autofill.ts           ← Injected into every page, detects login forms
```

### Auth Flow in Extension

1. First use: popup shows "Connect to Redandan" button
2. User clicks → opens `https://redandan.vercel.app/api/auth/extension-token` in a tab
3. App returns a long-lived token (stored in `chrome.storage.local`)
4. Extension uses this token as a `Bearer` header on all API calls
5. Token is refreshed by the background service worker before expiry

### Autofill Flow

1. Content script runs on every page load
2. Detects `<input type="password">` — signals a login form
3. Fetches `GET /api/credentials?site=<current-domain>` with the stored token
4. If matches found, shows a small Redandan icon inside the input field
5. User clicks icon → popup shows matched credentials → user clicks to fill

### Cross-Browser Notes

Firefox and Chrome both support Manifest V3 but with some differences:
- Firefox uses `browser.*` namespace (polyfilled with `webextension-polyfill`)
- `service_worker` in `background` is Chrome; Firefox uses `scripts` instead
- The `manifest.json` can include Firefox-specific keys under `browser_specific_settings`

Build with `esbuild` for fast, simple bundling. No webpack needed.

---

## Platform 3 — Mobile (Android)

**Stack:** Capacitor 6 wrapping the Next.js build
**Target:** Android (personal sideload)
**Status:** Phase 7

### How Capacitor Works

Capacitor is a bridge that runs a web app inside a native WebView and provides JavaScript APIs for native device features.

```
Next.js build (static export)
  → copied into Capacitor's www/ folder
  → Capacitor wraps it in an Android Activity
  → Native APIs available via @capacitor/* plugins
```

### Setup Steps (Phase 7)

1. Set `output: 'export'` in `next.config.ts` (static HTML/CSS/JS output)
2. Run `next build` → generates `out/` folder
3. Install Capacitor: `bun add @capacitor/core @capacitor/cli`
4. Initialize: `bun cap init redandan com.personal.redandan`
5. Set `webDir: 'out'` in `capacitor.config.ts`
6. Add Android: `bun cap add android`
7. Sync: `bun cap sync`
8. Open Android Studio: `bun cap open android`
9. Build APK → sideload onto phone

### Native Features Used

| Plugin | Purpose |
|---|---|
| `@capacitor/push-notifications` | Receive push notifications when app is closed |
| `@capacitor/local-notifications` | Schedule local reminders (e.g., deadline at 9am) |
| `@capacitor/haptics` | Subtle vibration feedback on task completion |
| `@capacitor/status-bar` | Match status bar to app theme |
| `@capacitor/app` | Handle back button, app state |

### API Connectivity

The mobile app calls the same Vercel API as the web app.
The `capacitor.config.ts` sets the server URL:
```ts
const config: CapacitorConfig = {
  appId: 'com.personal.redandan',
  appName: 'Redandan',
  webDir: 'out',
  server: {
    url: 'https://redandan.vercel.app',
    cleartext: false
  }
}
```

This means the app requires an internet connection. Offline support is not a Phase 7 goal.

### Static Export Constraints

When Next.js uses `output: 'export'`:
- API routes do **not** run in the static export
- All API calls go to the Vercel deployment (not local)
- Dynamic routes must have `generateStaticParams()` or use client-side fetching
- Server Components that fetch data must be converted to client components with `useEffect` + API calls, OR data is fetched server-side at build time (not ideal for dynamic personal data)

**Resolution:** For the mobile build, use a hybrid approach:
- Pages are static shells
- Data is fetched client-side via the API on mount
- This is acceptable for a personal app (not a public-facing performance-critical site)

---

## Platform Comparison

| Feature | Web | Extension | Mobile |
|---|---|---|---|
| Task management | Full | Quick-add only | Full |
| Password vault | Full | Autofill + add | View only |
| Calendar | Full | No | Full |
| Push notifications | Via Web Push | No | Native |
| Offline support | No | No | No |
| Auth method | Session cookie | Stored JWT token | Session cookie (WebView) |
