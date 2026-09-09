# Reliva — Documentation

> Personal Centralized Life & Tech Hub

This folder contains all technical documentation for the Reliva project.
Read this index first, then navigate to the relevant document.

---

## Documents

| File | Description |
|---|---|
| [overview.md](overview.md) | Vision, philosophy, and what Reliva is |
| [architecture.md](architecture.md) | System architecture, platform diagram, data flow |
| [tech-stack.md](tech-stack.md) | Full technology stack with versions and rationale |
| [technical-decisions.md](technical-decisions.md) | Architecture Decision Records (ADRs) |
| [project-structure.md](project-structure.md) | Annotated folder and file structure |
| [database.md](database.md) | MongoDB collections, schemas, and indexes |
| [api.md](api.md) | API route reference (endpoints, methods, shapes) |
| [security.md](security.md) | Security model, threat analysis, mitigations |
| [platforms.md](platforms.md) | Web, mobile, and browser extension strategy |
| [roadmap.md](roadmap.md) | 7-phase development plan |
| [environment.md](environment.md) | Environment variables reference |

---

## Quick Reference

**Current phase:** Phase 1 complete (foundation, auth, multi-user data model) — Phase 2 (Task
Tracker) in progress. See `roadmap.md`.

**Stack at a glance:**
- Backend: Go 1.25 + Gin, MongoDB (native driver, no ORM), custom JWT + TOTP auth — standalone,
  deployed independently of the frontend (Railway/Fly.io)
- Frontend: Next.js 16 + React 19 + TypeScript, pure client calling the Go API — no API routes, no
  DB access, no Auth.js
- Tailwind CSS v4 + shadcn/ui
- Deployed: `apps/web` on Vercel, `apps/server` on Railway/Fly.io as a Docker container

**Three platforms, one Go API:**
- Web → Next.js on Vercel
- Mobile → Capacitor (Phase 7, not started)
- Browser Extension → Manifest V3 (Phase 6, not started)

See `docs/architecture.md` and ADR-011 in `technical-decisions.md` for why the backend is a separate
Go server rather than Next.js API routes.

---

## Where to Start

If you're picking this up after a break, read in this order:

1. [overview.md](overview.md) — understand what you're building and why
2. [roadmap.md](roadmap.md) — find your current phase
3. [architecture.md](architecture.md) — orient yourself in the system
4. The relevant module doc for what you're working on

---

*Keep this documentation up to date as the project evolves.*
