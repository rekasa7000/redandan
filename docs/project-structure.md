# Project Structure

## Current State (Scaffold)

```
redandan/
  app/
    favicon.ico
    globals.css
    layout.tsx
    page.tsx
  components/
    ui/                   ← shadcn/ui components (generated)
    theme-provider.tsx
  hooks/                  ← empty
  lib/
    utils.ts              ← cn() utility
  public/
  docs/                   ← this folder
  .gitignore
  .prettierrc
  components.json         ← shadcn configuration
  eslint.config.mjs
  next.config.ts
  package.json
  postcss.config.mjs
  tsconfig.json
```

---

## Target State (Full Build)

```
redandan/
  ├── app/
  │   ├── api/                          ← All API route handlers
  │   │   ├── auth/
  │   │   │   └── [...nextauth]/
  │   │   │       └── route.ts          ← Auth.js endpoints
  │   │   ├── tasks/
  │   │   │   ├── route.ts              ← GET /api/tasks, POST /api/tasks
  │   │   │   └── [id]/
  │   │   │       └── route.ts          ← GET, PATCH, DELETE /api/tasks/:id
  │   │   ├── contexts/
  │   │   │   ├── route.ts
  │   │   │   └── [id]/route.ts
  │   │   ├── events/
  │   │   │   ├── route.ts
  │   │   │   └── [id]/route.ts
  │   │   ├── credentials/
  │   │   │   ├── route.ts              ← Returns ciphertext only, no decryption
  │   │   │   └── [id]/route.ts
  │   │   ├── notifications/
  │   │   │   ├── route.ts
  │   │   │   └── [id]/route.ts
  │   │   ├── push/
  │   │   │   ├── subscribe/route.ts    ← Save push subscription
  │   │   │   └── send/route.ts         ← Trigger push (internal)
  │   │   └── cron/
  │   │       └── notify/route.ts       ← Vercel Cron Job endpoint
  │   │
  │   ├── (auth)/                       ← Public routes (no auth required)
  │   │   └── login/
  │   │       └── page.tsx
  │   │
  │   ├── (app)/                        ← Protected routes (auth required)
  │   │   ├── layout.tsx                ← App shell: nav, sidebar
  │   │   ├── page.tsx                  ← Dashboard / Home
  │   │   ├── tasks/
  │   │   │   ├── page.tsx              ← Task list
  │   │   │   ├── new/page.tsx
  │   │   │   └── [id]/page.tsx
  │   │   ├── calendar/
  │   │   │   └── page.tsx
  │   │   ├── vault/
  │   │   │   ├── page.tsx              ← Credential list
  │   │   │   ├── new/page.tsx
  │   │   │   └── [id]/page.tsx
  │   │   └── settings/
  │   │       └── page.tsx
  │   │
  │   ├── favicon.ico
  │   ├── globals.css
  │   └── layout.tsx                    ← Root layout (providers, fonts)
  │
  ├── components/
  │   ├── ui/                           ← shadcn/ui primitives (owned, editable)
  │   │   ├── button.tsx
  │   │   ├── input.tsx
  │   │   ├── dialog.tsx
  │   │   └── ...
  │   ├── shared/                       ← App-wide layout components
  │   │   ├── navbar.tsx
  │   │   ├── sidebar.tsx
  │   │   ├── bottom-nav.tsx            ← Mobile bottom navigation
  │   │   └── notification-bell.tsx
  │   ├── tasks/                        ← Task-specific components
  │   │   ├── task-card.tsx
  │   │   ├── task-form.tsx
  │   │   ├── task-list.tsx
  │   │   └── task-filters.tsx
  │   ├── calendar/
  │   │   ├── calendar-view.tsx
  │   │   └── event-item.tsx
  │   ├── vault/                        ← Password vault components
  │   │   ├── credential-card.tsx
  │   │   ├── credential-form.tsx
  │   │   ├── master-password-gate.tsx  ← Unlock prompt
  │   │   └── password-generator.tsx
  │   └── dashboard/
  │       ├── today-widget.tsx
  │       ├── upcoming-widget.tsx
  │       └── quick-add.tsx
  │
  ├── lib/
  │   ├── db.ts                         ← MongoDB client singleton + collection getters
  │   ├── auth.ts                       ← Auth.js configuration
  │   ├── crypto.ts                     ← Web Crypto API helpers (PBKDF2, AES-GCM)
  │   ├── push.ts                       ← Web Push (VAPID) helpers
  │   ├── validations.ts                ← Zod schemas (shared client/server)
  │   ├── types.ts                      ← TypeScript interfaces for DB documents
  │   └── utils.ts                      ← cn() and other utilities
  │
  ├── hooks/
  │   ├── use-tasks.ts                  ← Task data fetching hook
  │   ├── use-vault.ts                  ← Vault state + crypto operations
  │   ├── use-master-key.ts             ← Master key session management
  │   └── use-notifications.ts          ← Push subscription management
  │
  ├── extension/                        ← Browser extension (separate build)
  │   ├── manifest.json
  │   ├── popup/
  │   │   ├── index.html
  │   │   └── popup.ts
  │   ├── background/
  │   │   └── service-worker.ts
  │   ├── content/
  │   │   └── autofill.ts
  │   └── icons/
  │       ├── icon-16.png
  │       ├── icon-48.png
  │       └── icon-128.png
  │
  ├── docs/                             ← This folder
  │   ├── README.md
  │   ├── overview.md
  │   ├── architecture.md
  │   ├── tech-stack.md
  │   ├── technical-decisions.md
  │   ├── project-structure.md
  │   ├── database.md
  │   ├── api.md
  │   ├── security.md
  │   ├── platforms.md
  │   ├── roadmap.md
  │   └── environment.md
  │
  ├── public/
  │   └── icons/                        ← PWA and favicon assets
  │
  ├── .env.local                        ← Local secrets (gitignored)
  ├── .env.example                      ← Example env file (committed)
  ├── .gitignore
  ├── .prettierrc
  ├── components.json                   ← shadcn configuration
  ├── eslint.config.mjs
  ├── middleware.ts                     ← Auth guard (edge)
  ├── next.config.ts
  ├── package.json
  ├── postcss.config.mjs
  ├── tsconfig.json
  └── vercel.json                       ← Cron job configuration
```

---

## Key File Responsibilities

| File | Responsibility |
|---|---|
| `lib/db.ts` | MongoDB connection singleton. Call `getDb()` to get the database instance. Never open multiple connections. |
| `lib/auth.ts` | Auth.js config: providers, callbacks, session strategy. Imported by the `[...nextauth]` route and `middleware.ts`. |
| `lib/crypto.ts` | All Web Crypto API operations: key derivation, encryption, decryption. Used only in client components and the extension. |
| `lib/types.ts` | TypeScript interfaces matching MongoDB document shapes. `Task`, `Context`, `Event`, `Credential`, `User`, `Notification`. |
| `lib/validations.ts` | Zod schemas for all entities. Used in both API routes (server validation) and forms (client validation). |
| `middleware.ts` | Runs at the edge on every request. Redirects unauthenticated users away from `(app)` routes. |
| `vercel.json` | Defines Vercel Cron Jobs. One job: daily at 8am, calls `/api/cron/notify`. |

---

## Naming Conventions

| Item | Convention | Example |
|---|---|---|
| Files | kebab-case | `task-card.tsx` |
| React components | PascalCase | `TaskCard` |
| Functions / variables | camelCase | `getTaskById` |
| API routes | plural noun | `/api/tasks`, `/api/contexts` |
| MongoDB collections | plural noun | `tasks`, `contexts` |
| Environment variables | SCREAMING_SNAKE_CASE | `MONGODB_URI` |
