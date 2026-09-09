# AI Agent Instructions

Rules and constraints for any AI agent or assistant working on this codebase.
Read this document in full before writing any code. These rules are not suggestions.

---

## Before You Start

1. Read `docs/overview.md` — understand what this app is and who it's for
2. Read `docs/architecture.md` — understand the system before touching it
3. Check `docs/roadmap.md` — identify the current phase and stay within it
4. Check `docs/technical-decisions.md` — do not re-litigate decided choices

---

## Git Rules

### Commit messages must be descriptive
Use conventional commit format: `type: short summary` followed by a detailed body.

Types: `feat`, `fix`, `docs`, `refactor`, `chore`, `style`, `test`

Good:
```
feat: add task creation API with Zod validation

Implements POST /api/tasks with full request body validation via Zod.
Inserts into the tasks collection using the native MongoDB driver.
Returns the created document with a 201 status.
```

Bad:
```
add task api
updated files
fix
```

### Never include AI co-author attribution in commits
Do not add `Co-Authored-By: Claude`, `Co-Authored-By: GitHub Copilot`, or any
AI co-author tag to commit messages.

### Commit granularity
- One logical change per commit
- Do not batch unrelated changes into one commit
- Do not commit incomplete work unless explicitly asked

---

## Design & UI Rules

### No gradients
Do not use CSS gradients anywhere in the UI. This includes:
- `bg-gradient-*` Tailwind classes
- `linear-gradient()` or `radial-gradient()` in inline styles or CSS
- Gradient borders or text gradients

Use flat, solid colors only. The UI should be clean and minimal.

### Use the design system — do not invent components
- Use existing shadcn/ui components from `components/ui/` before building anything new
- If a shadcn component does not exist yet, add it via `bunx shadcn add <component>`
- Do not write custom component primitives (buttons, inputs, dialogs) from scratch

### Color usage
- Use theme tokens (`bg-background`, `text-foreground`, `text-muted-foreground`, etc.)
- Use context colors (defined per context in the database) for context-specific badges
- Do not hardcode hex values in component files

### Mobile-first
- Write Tailwind classes mobile-first: base styles for small screens, `md:` and `lg:` for larger
- The app must be fully usable on a 375px wide screen
- Test layout at both small and large breakpoints before considering a UI task done

### No unnecessary animations
- Do not add decorative animations or transitions unless they serve a functional purpose
- `tw-animate-css` is available but should be used sparingly

---

## Code Rules

The backend (`apps/server`, Go) and frontend (`apps/web`, Next.js) are separate applications with
separate rules. **There is no `app/api/**` in `apps/web` — it does not exist and should not be
recreated.** All business logic and database access live exclusively in `apps/server`. See
`docs/architecture.md` and ADR-011 in `docs/technical-decisions.md` for why.

### No ORM — ever
Use only the native `go.mongodb.org/mongo-driver/v2` driver in `apps/server`. Do not install or
suggest an ODM. This is a firm architectural decision (see ADR-001).

### Go handler structure — mandatory pattern
Every handler in `apps/server/internal/handlers/*.go` follows this order (see any existing handler
for a concrete example):

```go
func (h *Handler) DoThing(c *gin.Context) {
    // 1. Get the caller's userID from context (set by RequireAuth/RequirePending middleware —
    //    the route registration in routes.go is what actually enforces auth, not this line)
    userID := c.GetString(middleware.ContextKeyUserID)
    oid, err := bson.ObjectIDFromHex(userID)
    if err != nil { c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"}); return }

    // 2. Bind + validate the request body (Gin `binding` tags)
    var body struct{ /* ... */ }
    if err := c.ShouldBindJSON(&body); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return
    }

    // 3. Business logic + DB access, always filtered by user_id
    // ...

    // 4. Return response
    c.JSON(http.StatusOK, result)
}
```

Every route must be registered under `RequireAuth`, `RequirePending`, or `RequireCron` in
`internal/routes/routes.go` — a route outside all three is a security bug. Every query on
user-owned data must filter by `user_id` (or `_id` + `user_id` for single-resource lookups) — this
is what makes the schema multi-user-safe (see ADR-012).

### TypeScript strictness (apps/web)
- All new code must be properly typed — no `any` unless absolutely unavoidable
- Document the reason in a comment if `any` is used
- Use the generated types in `apps/web/lib/types.gen.ts` (from `openapi.yaml`) for API shapes —
  don't hand-write duplicate interfaces for things the API already returns

### Frontend — pure client, no API routes, no DB access
- `apps/web` never touches MongoDB and never will. If a page needs data, it calls the Go API via
  `fetch` with `Authorization: Bearer <token>` (token from `lib/session.ts`)
- Every data-fetching page/component is a Client Component (`"use client"`) — there is no
  server-side data fetching to a database to fall back to
- Zod validation (planned for `lib/validations.ts`, Phase 2+) is for client-side form validation
  only — request validation happens server-side in the Go handler regardless

### Authentication — never bypass or weaken
- The app uses two-factor authentication: password + TOTP. Both factors are always required
- Do not add a path to bypass TOTP "for development" or "for testing"
- Do not create mock auth or skip TOTP in any code path, including scripts
- Auth is entirely in `apps/server` — `middleware.RequireAuth`/`RequirePending` in
  `internal/middleware/auth.go`. Never roll a separate auth check in a handler
- `apps/web` has no auth logic beyond storing/reading the token (`lib/session.ts`) and an edge
  presence-check (`middleware.ts`) — it never validates the JWT itself

### Standalone API principle
- The API (`apps/server`, `/api/v1/**`) must be client-agnostic. It must not return HTML,
  redirects, or client-specific responses
- Any data the web frontend needs must be available via the API, accessible to extension and
  mobile with the same bearer token mechanism — nothing web-only
- See ADR-009 and ADR-011 in `docs/technical-decisions.md`

### Password vault — client-side encryption only (Phase 4)
- Decryption logic must only exist in `apps/web/lib/crypto.ts` and the browser extension
- No Go handler may decrypt credential data — see `internal/handlers/credentials.go` for the
  current (correct) pattern of storing/returning ciphertext verbatim
- Never log credential plaintext, server or client
- Never pass the master password or derived CryptoKey to an API endpoint
- See `docs/security.md` for the full rule set

### Connection singleton (apps/server)
- Always use the `Handler.col()` helper (`internal/handlers/handler.go`) to get a collection —
  never instantiate a new `mongo.Client` inside a handler
- The singleton connection lives in `internal/db/db.go` (`sync.Once`) — don't add a second one

### No unnecessary abstractions
- Do not create utility functions, helpers, or wrappers for logic used only once
- Do not design for hypothetical future requirements
- Three similar lines of code is better than a premature abstraction

---

## What NOT to Do

- Do not install new dependencies without a clear reason — check if existing packages cover the need first
- Do not add features that are outside the current roadmap phase
- Do not refactor working code while implementing a feature — keep changes focused
- Do not add comments explaining what code does — only comment on *why* if the reasoning is non-obvious
- Do not create files unless they are directly needed for the task
- Do not push to remote unless explicitly instructed by the user
- Do not create pull requests unless explicitly instructed
- Do not run `git add .` or `git add -A` — stage specific files by name

---

## Package Manager

`apps/server` is a Go module — use `go mod tidy`, `go build`, `go run`, not bun. Everything else
(the workspace root and `apps/web`) uses **bun**. Always use:
```bash
bun install
bun add <package>
bun run <script>
bun dev
bun build
```

Never use `npm`, `yarn`, or `pnpm`.

---

## File Naming

| Item | Convention |
|---|---|
| Component files (`apps/web`) | `kebab-case.tsx` |
| Utility / lib files (`apps/web`) | `kebab-case.ts` |
| Go files (`apps/server`) | one file per resource in `internal/handlers/`, e.g. `tasks.go` |
| Type names | `PascalCase` |
| Function names (TS) | `camelCase` |
| Function names (Go) | `PascalCase` exported / `camelCase` unexported |
| Constants | `SCREAMING_SNAKE_CASE` |

---

## Asking vs. Acting

When in doubt about scope or approach, ask before acting. Specifically:
- If a task would require modifying more than 5 files, confirm the approach first
- If a task is ambiguous about which phase it belongs to, ask
- If a dependency needs to be added that is not already in `package.json`, ask first
- If destructive operations are needed (delete collection, reset data), always confirm
