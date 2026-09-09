<!-- BEGIN:nextjs-agent-rules -->
# This is NOT the Next.js you know

This version has breaking changes — APIs, conventions, and file structure may all differ from your training data. Read the relevant guide in `node_modules/next/dist/docs/` before writing any code. Heed deprecation notices.
<!-- END:nextjs-agent-rules -->

---

# Reliva — Agent Rules

Full agent documentation is in `docs/agents.md`. The rules below are mandatory and apply to every session.

---

## Orientation

Before writing any code, read:
- `docs/overview.md` — what this app is
- `docs/architecture.md` — how the system is structured
- `docs/roadmap.md` — current phase and scope
- `docs/technical-decisions.md` — decisions already made; do not re-open them

---

## Git

**Commits must be descriptive.** Use conventional commit format with a subject line and a body that explains what changed and why. One logical change per commit.

```
feat: add task creation API with Zod validation

Implements POST /api/tasks. Validates the request body with a Zod schema
before inserting into MongoDB. Returns the created task document as JSON.
```

**Never add AI co-author attribution.** Do not include `Co-Authored-By: Claude`,
`Co-Authored-By: GitHub Copilot`, or any AI attribution in commit messages.

**Never use `git add .` or `git add -A`.** Stage specific files by name.

**Never push or create pull requests** unless explicitly instructed.

---

## Design

- **No gradients.** Do not use `bg-gradient-*`, `linear-gradient()`, or any gradient syntax anywhere.
- **Flat, solid colors only.** Use theme tokens (`bg-background`, `text-foreground`, etc.).
- **Use shadcn/ui components** from `components/ui/`. Do not build primitive components from scratch.
- **Mobile-first.** Base Tailwind styles target small screens; use `md:` and `lg:` for larger.

---

## Code

`apps/server` (Go) owns all business logic and database access. `apps/web` (Next.js) is a pure
frontend — it has **no** API routes and **no** MongoDB access. There is no `app/api/**` in
`apps/web`; don't recreate one. See `docs/architecture.md` and ADR-011 in `docs/technical-decisions.md`.

- **No ORM.** `apps/server` uses only the native `mongo-driver/v2` Go driver. No ODM. Ever.
- **No `any`** in TypeScript. All code must be properly typed. If `any` is unavoidable, comment the reason.
- **Auth lives in `apps/server` only.** Every protected Go route is registered under
  `middleware.RequireAuth`/`RequirePending`/`RequireCron` in `internal/routes/routes.go`. Every
  client — web, extension, mobile — authenticates with the same bearer JWT; there is no
  cookie-session code path to keep in sync.
- **Every query is scoped by `user_id`** (from the JWT). The data model is multi-user by design —
  see ADR-012 — even though only one operator uses it today. Don't special-case "the one user."
- **TOTP is mandatory.** Never add a bypass, skip, or mock path for the two-factor auth flow. Both factors — password and TOTP — are always required.
- **Standalone API.** The Go API must work identically for all clients. Do not add logic that only works for the web client and not for bearer token callers.
- **Password vault is client-side only.** No Go handler may decrypt credentials. Master password never leaves the browser. See `docs/security.md`.
- **Use `Handler.col()`** (`apps/server/internal/handlers/handler.go`) to get a Mongo collection — never instantiate a new client in a handler.
- **No unnecessary abstractions.** Do not create helpers for one-time use.

---

## Package Manager

`apps/server` is a Go module — use `go mod tidy` / `go build` / `go run`, not bun.
Everywhere else (workspace root, `apps/web`), use **bun** exclusively. Never use npm, yarn, or pnpm.

```bash
bun install / bun add / bun dev / bun build
```

---

## Scope

- Stay within the current roadmap phase (check `docs/roadmap.md`)
- Do not add features outside the phase in progress
- Do not refactor working code while implementing a feature
- If a task touches more than 5 files or is ambiguous, confirm the approach before acting
