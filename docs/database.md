# Database

## Overview

- **Database:** MongoDB 7.x
- **Host:** MongoDB Atlas (free M0 tier)
- **Driver:** `mongodb` (native Node.js driver, no ORM)
- **Connection:** Singleton in `lib/db.ts`, pooled across serverless function invocations

---

## Connection Pattern

```ts
// lib/db.ts
import { MongoClient, Db } from 'mongodb'

const uri = process.env.MONGODB_URI!
const options = {}

let client: MongoClient
let db: Db

if (process.env.NODE_ENV === 'development') {
  // In dev, reuse the client across hot reloads
  if (!(global as any)._mongoClient) {
    (global as any)._mongoClient = new MongoClient(uri, options)
  }
  client = (global as any)._mongoClient
} else {
  client = new MongoClient(uri, options)
}

export async function getDb(): Promise<Db> {
  if (!db) {
    await client.connect()
    db = client.db('reliva')
  }
  return db
}
```

Usage in a route handler:
```ts
const db = await getDb()
const tasks = await db.collection('tasks').find({ status: 'todo' }).toArray()
```

---

## Collections

### `users`

Single document. The one registered user.

```ts
interface User {
  _id: ObjectId
  username: string
  passwordHash: string          // bcrypt hash
  pushSubscriptions: PushSubscription[]
  createdAt: Date
}
```

Indexes:
- `{ username: 1 }` unique

---

### `contexts`

Categories / life areas. Tasks belong to a context.

```ts
interface Context {
  _id: ObjectId
  name: string                  // e.g. "Job 1 — Acme Corp"
  slug: string                  // e.g. "job-1-acme"
  color: string                 // hex color for UI display
  icon: string                  // lucide icon name
  type: 'work' | 'personal' | 'health' | 'finance' | 'travel' | 'custom'
  order: number                 // display order
}
```

Indexes:
- `{ slug: 1 }` unique

Seeded contexts:
```json
[
  { "name": "Personal", "slug": "personal", "type": "personal", "color": "#6366f1", "icon": "user" },
  { "name": "Health", "slug": "health", "type": "health", "color": "#22c55e", "icon": "heart" },
  { "name": "Finance", "slug": "finance", "type": "finance", "color": "#f59e0b", "icon": "wallet" },
  { "name": "Travel", "slug": "travel", "type": "travel", "color": "#0ea5e9", "icon": "plane" }
]
```
Work contexts are added manually per job.

---

### `tasks`

The core collection. Every task lives here.

```ts
interface Task {
  _id: ObjectId
  title: string
  description?: string
  contextId: ObjectId           // reference to contexts._id
  priority: 'low' | 'medium' | 'high' | 'urgent'
  status: 'todo' | 'in_progress' | 'done' | 'archived'
  deadline?: Date
  reminderAt?: Date
  recurrence?: 'none' | 'daily' | 'weekly' | 'monthly'
  tags: string[]
  notes?: string
  createdAt: Date
  updatedAt: Date
}
```

Indexes:
- `{ status: 1, deadline: 1 }` — for dashboard queries (active tasks sorted by deadline)
- `{ contextId: 1, status: 1 }` — for context-filtered views
- `{ deadline: 1 }` — for cron notification queries
- `{ tags: 1 }` — for tag filtering

Common queries:
```ts
// Tasks due today
const today = new Date()
today.setHours(0, 0, 0, 0)
const tomorrow = new Date(today)
tomorrow.setDate(tomorrow.getDate() + 1)

db.collection('tasks').find({
  deadline: { $gte: today, $lt: tomorrow },
  status: { $in: ['todo', 'in_progress'] }
})

// Overdue tasks
db.collection('tasks').find({
  deadline: { $lt: today },
  status: { $in: ['todo', 'in_progress'] }
})

// Tasks by context
db.collection('tasks').find({
  contextId: new ObjectId(contextId),
  status: { $ne: 'archived' }
}).sort({ deadline: 1, priority: -1 })
```

---

### `events`

Calendar entries — things with a date but not necessarily a task.

```ts
interface Event {
  _id: ObjectId
  title: string
  type: 'payroll' | 'vacation' | 'deadline' | 'appointment' | 'custom'
  contextId?: ObjectId          // optional: which job/area this belongs to
  date: Date                    // start date
  endDate?: Date                // for multi-day events (vacations)
  allDay: boolean
  recurrence: 'none' | 'monthly' | 'annually'
  notes?: string
  createdAt: Date
}
```

Indexes:
- `{ date: 1 }` — for calendar range queries
- `{ type: 1, date: 1 }` — for cron: find upcoming payroll events

Common queries:
```ts
// Events in a date range (calendar view)
db.collection('events').find({
  date: { $gte: rangeStart, $lte: rangeEnd }
}).sort({ date: 1 })

// Upcoming payroll in next 2 days
db.collection('events').find({
  type: 'payroll',
  date: { $gte: today, $lte: twoDaysFromNow }
})
```

---

### `credentials`

Encrypted password vault entries.

```ts
interface Credential {
  _id: ObjectId
  site: string                  // display name, e.g. "GitHub"
  siteUrl: string               // e.g. "https://github.com"
  username: string              // plaintext — not sensitive
  encryptedPassword: string     // base64 AES-GCM ciphertext
  iv: string                    // base64 initialization vector
  salt: string                  // base64 PBKDF2 salt
  encryptedNotes?: string       // base64 ciphertext (optional)
  notesIv?: string
  tags: string[]
  lastModified: Date
  createdAt: Date
}
```

Indexes:
- `{ siteUrl: 1 }` — for extension domain lookup
- `{ tags: 1 }` — for tag filtering

Notes:
- The server returns `encryptedPassword`, `iv`, `salt` as-is
- The client decrypts using the master-derived key
- `username` is stored plaintext so the extension can display it without decryption
- Never add a server-side function to decrypt these — decryption is client-only

---

### `notifications`

Log of sent notifications and their read status.

```ts
interface Notification {
  _id: ObjectId
  type: 'task_due' | 'task_overdue' | 'event_reminder' | 'overdue_digest'
  refId?: ObjectId              // the task or event that triggered this
  refType?: 'task' | 'event'
  message: string
  sentAt: Date
  read: boolean
}
```

Indexes:
- `{ read: 1, sentAt: -1 }` — for unread badge and notification list

---

## Data Integrity Rules

Since there is no ORM enforcing relationships, these rules must be handled in application code:

| Rule | Where enforced |
|---|---|
| `tasks.contextId` must reference a valid context | Validate in API route before insert |
| `events.contextId` must reference a valid context if provided | Validate in API route |
| When a context is deleted, update its tasks to a "general" context | `DELETE /api/contexts/:id` handler |
| Credentials must have `encryptedPassword`, `iv`, and `salt` | Zod schema on the API route |

---

## Seeding

On first deploy, run the seed script to create the admin user and default contexts:

```ts
// scripts/seed.ts
import { getDb } from '../lib/db'
import bcrypt from 'bcrypt'

const db = await getDb()

await db.collection('users').insertOne({
  username: process.env.SEED_USERNAME,
  passwordHash: await bcrypt.hash(process.env.SEED_PASSWORD, 12),
  pushSubscriptions: [],
  createdAt: new Date()
})

// Insert default contexts...
```

Run with: `bun run scripts/seed.ts`
