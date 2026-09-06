# @redandan/extension

Browser extension for Chrome and Firefox. Manifest V3.

**Status:** Phase 6 — not yet implemented.

See `docs/platforms.md` for the full extension architecture and plan.

## What it will do

- Password autofill on login forms (fetches from the Redandan API)
- Quick task creation from any tab
- Notification badge showing overdue task count

## Tech

- Manifest V3 (Chrome + Firefox compatible)
- TypeScript compiled with esbuild
- Communicates with the Vercel-hosted API using bearer tokens
