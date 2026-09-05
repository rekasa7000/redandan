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
```
```
updated files
```
```
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

### No ORM — ever
Use only the native `mongodb` driver. Do not install or suggest Mongoose, Prisma, Drizzle,
or any other ORM or ODM. This is a firm architectural decision (see ADR-001).

### TypeScript strictness
- All new code must be properly typed — no `any` unless absolutely unavoidable
- Document the reason in a comment if `any` is used
- Define document shapes as TypeScript interfaces in `lib/types.ts`

### Validation belongs in two places only
- **Server:** Zod schema validation at the top of every API route handler, before any DB call
- **Client:** The same Zod schema reused in form validation
- Do not validate in the database layer or in component logic

### Password vault — client-side encryption only
- Decryption logic must only exist in `lib/crypto.ts` and the browser extension
- No server-side function may decrypt credential data
- Never log credential plaintext in the server or the browser console
- Never pass the master password or derived CryptoKey to an API endpoint
- See `docs/security.md` for the full rule set

### Connection singleton
- Always use `getDb()` from `lib/db.ts` — never instantiate `MongoClient` directly in a route
- Never use `client.db()` outside of `lib/db.ts`

### API route structure
- Every route handler must check the session first, before any other logic
- Return errors using `Response.json({ error: '...' }, { status: ... })`
- Use the HTTP status codes defined in `docs/api.md`

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

This project uses **bun**. Always use:
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
| Component files | `kebab-case.tsx` |
| Utility / lib files | `kebab-case.ts` |
| API route files | `route.ts` (fixed by Next.js) |
| Type names | `PascalCase` |
| Function names | `camelCase` |
| Constants | `SCREAMING_SNAKE_CASE` |

---

## Asking vs. Acting

When in doubt about scope or approach, ask before acting. Specifically:
- If a task would require modifying more than 5 files, confirm the approach first
- If a task is ambiguous about which phase it belongs to, ask
- If a dependency needs to be added that is not already in `package.json`, ask first
- If destructive operations are needed (delete collection, reset data), always confirm
