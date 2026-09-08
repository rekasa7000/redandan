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

**Current phase:** Phase 1 — Foundation

**Stack at a glance:**
- Next.js 16 + React 19 + TypeScript
- MongoDB (native driver, no ORM)
- Auth.js (NextAuth v5)
- Tailwind CSS v4 + shadcn/ui
- Deployed on Vercel

**Three platforms, one codebase:**
- Web → Next.js on Vercel
- Mobile → Capacitor (Phase 7)
- Browser Extension → Manifest V3 (Phase 6)

---

## Where to Start

If you're picking this up after a break, read in this order:

1. [overview.md](overview.md) — understand what you're building and why
2. [roadmap.md](roadmap.md) — find your current phase
3. [architecture.md](architecture.md) — orient yourself in the system
4. The relevant module doc for what you're working on

---

*Keep this documentation up to date as the project evolves.*
