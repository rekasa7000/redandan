# Tech Stack

## Summary Table

| Layer | Technology | Version | Role |
|---|---|---|---|
| Framework | Next.js | 16.x | Full-stack: UI + API |
| Language | TypeScript | 5.x | Type safety across the whole codebase |
| Runtime | Node.js | 20.x | Server-side execution (Vercel) |
| Database | MongoDB | 7.x (Atlas) | Primary data store |
| DB Driver | mongodb (native) | 6.x | Direct queries, no ORM |
| UI Library | shadcn/ui | latest | Component library on top of Radix |
| Styling | Tailwind CSS | 4.x | Utility-first CSS |
| Auth | Auth.js (NextAuth) | v5 | Session management, credential provider |
| Validation | Zod | 3.x | Schema validation on client and server |
| Icons | Lucide React | latest | Icon set |
| Crypto | Web Crypto API | native | Client-side AES-GCM encryption |
| Push | web-push | latest | VAPID-based Web Push notifications |
| Package Manager | bun | latest | Fast installs and script runner |
| Linting | ESLint | 9.x | Code quality |
| Formatting | Prettier | 3.x | Code style |
| Deployment | Vercel | — | Hosting, Cron Jobs, Edge Network |
| Database Host | MongoDB Atlas | — | Cloud MongoDB, free M0 tier |
| Mobile | Capacitor | 6.x | Web-to-native wrapper (Phase 7) |
| Extension | Manifest V3 | — | Chrome + Firefox compatible (Phase 6) |

---

## Framework — Next.js 16 (App Router)

Next.js was chosen because it handles both the frontend and the backend in one project. The App Router (introduced in Next.js 13, now mature) enables:

- **React Server Components** — fetch data on the server, ship less JS to the client
- **Route Handlers** — the API backend, replacing Express entirely
- **Middleware** — runs at the edge for auth checks before any page or API route
- **File-based routing** — predictable, organized structure

There is no separate Express backend. No separate Fastify server. One `next dev` runs everything.

---

## Database — MongoDB (native driver)

MongoDB was chosen for its schema flexibility — tasks, events, and credentials all have different shapes and optional fields, which maps naturally to documents.

The **native `mongodb` driver** is used directly. No Mongoose, no Prisma, no Drizzle. This means:
- Queries are written in TypeScript against the driver's typed API
- No model abstraction layer between you and the database
- Full control over indexes, aggregation pipelines, and projections
- TypeScript interfaces define the document shapes

**MongoDB Atlas** free tier (M0) is sufficient: 512MB storage, shared cluster. More than enough for a personal app with a single user.

---

## Auth — Auth.js (NextAuth v5)

Auth.js provides:
- **Credential provider** — username/password login backed by MongoDB
- **JWT sessions** — stateless, works well with Vercel's serverless environment
- **Middleware integration** — `middleware.ts` uses Auth.js to protect all `(app)` routes
- **TypeScript-first** — fully typed session object

No social login is needed for a personal app. The single user is seeded into the database on first deploy.

---

## UI — shadcn/ui + Tailwind CSS v4

shadcn/ui is not a component library you install from npm — it's a collection of components you copy into your codebase and own. This means:
- Full control over styling and behavior
- No versioning conflicts with the component library
- Components are already in `components/ui/` and can be customized freely

Tailwind CSS v4 (the version already configured) introduces a new Vite-based engine and CSS-first configuration (no `tailwind.config.js` needed).

---

## Validation — Zod

Zod is used in two places:
1. **Server-side** — validate request bodies in API Route Handlers before touching the database
2. **Client-side** — validate form input before submitting (shared schemas in `lib/validations.ts`)

Using the same schema in both places eliminates duplication and ensures consistency.

---

## Cryptography — Web Crypto API

The password vault uses the browser-native Web Crypto API for all encryption. No third-party crypto library is needed.

The flow:
1. User enters master password
2. A `CryptoKey` is derived using **PBKDF2** with a random salt
3. Each credential is encrypted with **AES-GCM** (256-bit key, random IV per item)
4. Only the ciphertext, IV, and salt are stored in MongoDB
5. Decryption happens entirely in the browser — the server never sees plaintext

This approach means the server is useless to an attacker who only has the database.

---

## Push Notifications — web-push (VAPID)

Web Push notifications allow the browser (and PWA) to receive notifications even when the app is not open.

- VAPID keys identify the server to the push service
- Vercel Cron Jobs trigger the notification logic on a schedule
- Subscriptions are stored per-user in the `users` collection
- On mobile (Capacitor), native push replaces web push (Phase 7)

---

## Package Manager — bun

The project uses bun (evidenced by `bun.lock`). Use `bun install`, `bun dev`, `bun build`, etc. Do not switch to npm or yarn.

---

## Deployment — Vercel

Vercel is the deployment target. Key features used:
- **Auto-deploy from GitHub** — push to `main`, it deploys
- **Environment Variables** — set in the Vercel dashboard, available at runtime
- **Vercel Cron Jobs** — defined in `vercel.json`, used to trigger scheduled notifications
- **Edge Middleware** — `middleware.ts` runs at the edge for fast auth checks

The MongoDB Atlas connection string is stored as `MONGODB_URI` in Vercel's environment variables.
