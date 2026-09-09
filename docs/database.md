# Database

## Overview

- **Database:** MongoDB 8.x
- **Host:** Docker (`mongo:8`) locally via `infra/docker-compose.yml`; MongoDB Atlas or any reachable
  MongoDB instance in production, via `MONGO_URI`
- **Driver:** `go.mongodb.org/mongo-driver/v2` (native Go driver, no ORM)
- **Connection:** Singleton in `apps/server/internal/db/db.go`, connected once via `sync.Once` and
  reused across every request
- **Owner:** Only `apps/server` ever talks to MongoDB. No other client (web, extension, mobile)
  connects to the database directly

---

## Connection Pattern

```go
// internal/db/db.go
var (
    client  *mongo.Client
    once    sync.Once
    initErr error
)

func Connect(uri string) (*mongo.Client, error) {
    once.Do(func() {
        c, err := mongo.Connect(options.Client().ApplyURI(uri))
        if err != nil { initErr = err; return }
        if err = c.Ping(context.Background(), nil); err != nil { initErr = err; return }
        client = c
    })
    return client, initErr
}
```

Handlers get a collection via a shared helper on the `Handler` struct
(`apps/server/internal/handlers/handler.go`):
```go
func (h *Handler) col(name string) *mongo.Collection {
    return h.db.Collection(name)
}
```

---

## Collections

All document structs live in `apps/server/internal/models/models.go`. Field names below are the
`bson` tags actually stored — every collection except `users` includes a `user_id` field, and every
handler query filters by it (extracted from the JWT's `sub` claim). This is what makes the schema
multi-user-safe: nothing hardcodes "the one user."

### `users`

```go
type User struct {
    ID                bson.ObjectID
    Email             string
    PasswordHash      string    // bcrypt
    TOTPSecret        string
    TOTPEnabled       bool
    TOTPPendingSecret string    // set during setup, cleared on confirm
    BackupCodes       []string  // bcrypt hashes, consumed on use
    CreatedAt         time.Time
}
```

No unique index is currently created on `email` in code — uniqueness is enforced by the seed
script's per-email existence check, not a database constraint. Multiple users are supported; there
is no self-serve signup endpoint, only the seed command (see `docs/environment.md`).

---

### `push_subscriptions`

```go
type PushSubscription struct {
    ID        bson.ObjectID
    UserID    bson.ObjectID
    Endpoint  string
    P256dh    string
    Auth      string
    CreatedAt time.Time
}
```

Upserted by `(user_id, endpoint)` on `POST /api/v1/push/subscribe` — re-subscribing the same
browser updates the existing record instead of duplicating it.

---

### `contexts`

Life areas / jobs. Tasks and (optionally) events belong to a context.

```go
type Context struct {
    ID          bson.ObjectID
    UserID      bson.ObjectID
    Name        string
    Slug        string    // only ever set by the seed script today
    Color       string
    Icon        string
    Type        string    // work | personal | health | finance | travel | custom — seed-only today
    Description string
    Order       int       // seed-only today
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

`POST /api/v1/contexts` only accepts `name`, `description`, `color`, `icon` — `slug`, `type`, and
`order` are set by `cmd/seed/main.go` when it creates the default contexts, but the handler doesn't
expose them yet.

Seeded per account (`cmd/seed/main.go`):
```json
[
  { "name": "Personal", "slug": "personal", "type": "personal", "color": "#6366f1", "icon": "user",      "order": 0 },
  { "name": "Work",     "slug": "work",     "type": "work",     "color": "#f59e0b", "icon": "briefcase", "order": 1 },
  { "name": "Health",   "slug": "health",   "type": "health",   "color": "#10b981", "icon": "heart",     "order": 2 }
]
```

---

### `tasks`

```go
type Task struct {
    ID          bson.ObjectID
    UserID      bson.ObjectID
    ContextID   bson.ObjectID
    Title       string
    Description string
    Priority    string     // free-form; no enum enforced in code today
    Status      string     // created as "pending"; "done" stamps CompletedAt
    DueDate     *time.Time
    ReminderAt  *time.Time
    Recurrence  string     // none | daily | weekly | monthly (not yet enforced)
    Tags        []string
    Notes       string
    CompletedAt *time.Time
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

Note: `CreateTask` sets the initial `status` to `"pending"`. Frontend/roadmap docs describe a
`todo | in_progress | done | archived` lifecycle — that enum isn't enforced server-side yet; `status`
is stored as whatever string the client sends, except on creation.

Common query patterns (all Go, run inside handlers, always filtered by `user_id`):
```go
// Tasks for a user, optionally filtered by status/context, sorted by due_date
filter := bson.M{"user_id": oid}
if status != "" { filter["status"] = status }
if contextID != "" { filter["context_id"] = cid }
h.col("tasks").Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "due_date", Value: 1}}))

// Cron: tasks due today across all users, not done/cancelled
bson.M{
    "status":   bson.M{"$nin": []string{"done", "cancelled"}},
    "due_date": bson.M{"$gte": startOfDay, "$lt": endOfDay},
}
```

No indexes are explicitly created in code yet — queries currently rely on collection scans plus
whatever default `_id` index MongoDB provides. Adding `{user_id: 1, status: 1}` and
`{user_id: 1, due_date: 1}` indexes is worth doing before real data volume arrives.

---

### `events`

```go
type Event struct {
    ID          bson.ObjectID
    UserID      bson.ObjectID
    ContextID   *bson.ObjectID  // optional
    Title       string
    Type        string          // payroll | vacation | deadline | appointment | custom (not enforced)
    Description string
    StartTime   time.Time
    EndTime     *time.Time
    AllDay      bool
    Recurrence  string          // none | monthly | annually (not enforced)
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

`ListEvents` returns all of the caller's events sorted by `start_time` ascending — there's no date
range filter on the query params yet (Phase 3 work).

---

### `credentials`

Encrypted password vault entries. The server stores and returns ciphertext only.

```go
type Credential struct {
    ID            bson.ObjectID
    UserID        bson.ObjectID
    Site          string
    SiteURL       string
    Username      string    // plaintext — not sensitive
    EncryptedData string    // base64 AES-GCM ciphertext
    IV            string    // base64 initialization vector
    Salt          string    // base64 PBKDF2 salt
    Notes         string
    Tags          []string
    CreatedAt     time.Time
    UpdatedAt     time.Time
}
```

Never add a server-side function that decrypts `EncryptedData`. Decryption is client-only (see
`docs/security.md`).

---

### `notifications`

In-app notification log.

```go
type Notification struct {
    ID        bson.ObjectID
    UserID    bson.ObjectID
    Title     string
    Body      string
    Read      bool
    CreatedAt time.Time
}
```

Populated by `POST /api/v1/cron/notify` (one document per task due today) and read via
`GET /api/v1/notifications` (50 most recent, newest first).

---

## Data Integrity Rules

There is no ORM enforcing relationships or cross-document validation. Current state, honestly:

| Rule | Enforced? |
|---|---|
| Every query is scoped to the caller's `user_id` | Yes — in every handler, by hand |
| `tasks.context_id` references a valid context | No — not validated on insert |
| Deleting a context reassigns/orphans its tasks | No — `DeleteContext` does not touch `tasks` |
| Credentials always have `encrypted_data`, `iv`, `salt` | Yes — `binding:"required"` on the Gin struct |

These are worth tightening as Phase 2+ work lands, not urgent for Phase 1.

---

## Seeding

Accounts are created by the Go seed command (`apps/server/cmd/seed/main.go`), not a signup endpoint.
It checks for an existing user **by the specific email**, not "does any user exist" — so it can be
run multiple times to create additional accounts (see ADR-013 in `docs/technical-decisions.md`).

```bash
cd infra
# set SEED_EMAIL / SEED_PASSWORD in .env
docker compose run --rm seed
```

Required env vars: `MONGO_URI`, `SEED_EMAIL`, `SEED_PASSWORD`. Optional: `DB_NAME` (default `reliva`).

Each run creates the user (bcrypt-hashed password, TOTP disabled) plus the three default contexts
listed above.

There is also a legacy `scripts/seed.ts` (Bun + Node `mongodb` driver) at the monorepo root, written
before the Go seed command existed. It's not wired into any `package.json` script and duplicates
`cmd/seed/main.go` — treat the Go seed command as the source of truth.
