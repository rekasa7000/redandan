# API Reference

All routes are served by the Go + Gin standalone server (`apps/server`), mounted under `/api/v1`.
There are no Next.js API routes — `apps/web` calls these endpoints exactly like any other client.
All responses are `Content-Type: application/json`. The authoritative machine-readable contract is
`openapi.yaml` at the monorepo root (used to generate `apps/web/lib/types.gen.ts`).

Base URL:
- Local: `http://localhost:8080`
- Production: `https://<your-server>.railway.app`

## Authentication

Every protected route requires a bearer JWT issued by the Go server:

| Header | Value |
|---|---|
| `Authorization` | `Bearer <jwt>` |

There is no session cookie on the server side. The web frontend stores its token in a browser
cookie purely for its own edge-middleware check, then attaches it as an `Authorization` header on
every request like any other client (see `docs/architecture.md`).

Two token types exist, both signed with `JWT_SECRET`:

| Type | Issued when | Expiry | Valid on |
|---|---|---|---|
| `pending` | After step 1 (password), if TOTP is enabled | 5 minutes | `/auth/totp/validate`, `/auth/backup-code` only |
| `access` | After step 2 (TOTP/backup code), or after step 1 if TOTP isn't set up yet | 30 days | Every other protected route |

A route protected by `RequireAuth` rejects `pending` tokens and vice versa.

---

## Auth Endpoints

### `POST /api/v1/auth/login`
Step 1 of login. Validates email + password.

Request body:
```json
{ "email": "string", "password": "string" }
```

Response if TOTP is enabled — `200`:
```json
{ "pending_token": "<jwt>", "require_totp": true }
```

Response if TOTP is **not yet** set up — `200` (access token issued directly so the client can
complete setup):
```json
{ "token": "<jwt>", "require_totp": false }
```

Response on failure: `401 { "error": "invalid credentials" }`

---

### `POST /api/v1/auth/totp/validate`
Step 2 of login. Requires `Authorization: Bearer <pending token>`.

Request body:
```json
{ "code": "123456" }
```

Response: `200 { "token": "<access jwt>" }`
Response on failure: `401 { "error": "invalid TOTP code" }`

---

### `POST /api/v1/auth/backup-code`
Alternative to TOTP during login step 2. Requires `Authorization: Bearer <pending token>`.
Burns the code on success — it cannot be reused.

Request body:
```json
{ "code": "a1b2c3d4e5f6" }
```

Response: `200 { "token": "<access jwt>" }`
Response on failure: `401 { "error": "invalid backup code" }`

---

### `GET /api/v1/auth/me`
Requires `Authorization: Bearer <access token>`.

Response:
```json
{ "email": "string", "totp_enabled": true }
```

---

### `GET /api/v1/auth/totp/setup`
Generates a new TOTP secret (stored as `totp_pending_secret` until confirmed) and a QR code.
Requires an access token.

Response:
```json
{
  "qr": "data:image/png;base64,...",
  "secret": "BASE32SECRET",
  "uri": "otpauth://totp/Reliva:you@example.com?secret=...&issuer=Reliva"
}
```

### `POST /api/v1/auth/totp/confirm`
Confirms setup by verifying the first code against the pending secret. Activates TOTP and generates
8 single-use backup codes (shown once). Requires an access token.

Request body:
```json
{ "code": "123456" }
```

Response: `200 { "message": "TOTP enabled", "backup_codes": ["...", "..."] }`
Response on failure: `401 { "error": "invalid TOTP code" }`

---

### `POST /api/v1/auth/change-password`
Requires an access token.

Request body:
```json
{ "current_password": "string", "new_password": "string (min 8 chars)" }
```

Response: `200 { "message": "password updated" }`

---

### `POST /api/v1/auth/forgot-password`
Public. Resets the password using a backup code instead of email OTP — the backup code proves
identity and is burned on success.

Request body:
```json
{ "email": "string", "backup_code": "string", "new_password": "string (min 8 chars)" }
```

Response: `200 { "message": "password reset — log in with your new password" }`
Response on failure: `401 { "error": "invalid credentials" }` (same error whether the email exists
or the code is wrong — prevents account enumeration)

---

### `POST /api/v1/auth/logout`
Stateless — the client discards the token locally. Always returns `200 { "message": "logged out" }`.

---

## Tasks

### `GET /api/v1/tasks`
Requires an access token. Returns tasks belonging to the caller, sorted by `due_date` ascending.

Query params:
| Param | Description |
|---|---|
| `status` | Exact match, e.g. `pending`, `done` |
| `context_id` | Filter by context ObjectID |

Response: `200` — a JSON array of task documents (see `docs/database.md` for shape).

### `POST /api/v1/tasks`
```json
{
  "title": "string (required)",
  "description": "string",
  "context_id": "ObjectId string",
  "priority": "string",
  "due_date": "ISO date string",
  "tags": ["string"]
}
```
New tasks are created with `status: "pending"`. Response: `201` with the created document.

### `GET /api/v1/tasks/:id`
`200` with the task, or `404` if it doesn't exist or belongs to another user.

### `PATCH /api/v1/tasks/:id`
Partial update — send only the fields to change (`title`, `description`, `status`, `priority`,
`due_date`, `tags`, `context_id`). Setting `status: "done"` stamps `completed_at`.
Response: `200` with the updated document.

### `DELETE /api/v1/tasks/:id`
Response: `204` (no body). `404` if not found.

---

## Contexts

### `GET /api/v1/contexts`
Returns all contexts for the caller.

### `POST /api/v1/contexts`
```json
{
  "name": "string (required)",
  "description": "string",
  "color": "#hexcode",
  "icon": "string"
}
```
Note: the API does not currently accept `slug`, `type`, or `order` — those fields are only ever set
by the seed script's default contexts (Personal / Work / Health). Response: `201`.

### `GET /api/v1/contexts/:id`
### `PATCH /api/v1/contexts/:id`
Same partial-update pattern as tasks, over `name`, `description`, `color`, `icon`.
### `DELETE /api/v1/contexts/:id`

---

## Events

### `GET /api/v1/events`
Returns all events for the caller, sorted by `start_time` ascending.

### `POST /api/v1/events`
```json
{
  "title": "string (required)",
  "description": "string",
  "start_time": "ISO date string (required)",
  "end_time": "ISO date string",
  "all_day": false,
  "context_id": "ObjectId string"
}
```

### `GET /api/v1/events/:id`
### `PATCH /api/v1/events/:id`
### `DELETE /api/v1/events/:id`

---

## Credentials (Password Vault)

**The server never decrypts credential data.** `encrypted_data`, `iv`, and `salt` are stored and
returned verbatim — encryption/decryption happens entirely client-side.

### `GET /api/v1/credentials`
Response:
```json
[
  {
    "id": "...",
    "site": "GitHub",
    "site_url": "",
    "username": "myusername",
    "encrypted_data": "base64...",
    "iv": "base64...",
    "salt": "base64...",
    "notes": "",
    "tags": [],
    "created_at": "...",
    "updated_at": "..."
  }
]
```

### `POST /api/v1/credentials`
```json
{
  "site": "string (required)",
  "username": "string",
  "encrypted_data": "base64 ciphertext (required)",
  "iv": "base64 (required)",
  "salt": "base64 (required)",
  "notes": "string"
}
```

### `GET /api/v1/credentials/:id`
### `PATCH /api/v1/credentials/:id`
### `DELETE /api/v1/credentials/:id`

---

## Notifications

### `GET /api/v1/notifications`
Returns the 50 most recent notifications for the caller, newest first.

### `PATCH /api/v1/notifications/:id`
Marks a notification as read. No request body needed.
Response: `200 { "message": "marked as read" }`

---

## Push

### `POST /api/v1/push/subscribe`
Registers (or upserts, by endpoint) a Web Push subscription for the caller.

Request body — the raw `PushSubscription` shape from
`navigator.serviceWorker.pushManager.subscribe()`:
```json
{
  "endpoint": "string (required)",
  "keys": { "p256dh": "string (required)", "auth": "string (required)" }
}
```

Response: `201 { "message": "subscribed" }`

---

## Cron

### `POST /api/v1/cron/notify`
Not user-facing. Protected by `RequireCron` — a distinct check against `CRON_SECRET`, not a user
JWT.

Required header: `Authorization: Bearer <CRON_SECRET>`

Logic: finds all tasks due today (across all users) that aren't `done`/`cancelled`, and inserts one
notification document per task.

Response: `200 { "message": "notifications created", "created": 5 }`
Response if the secret is wrong: `403 { "error": "forbidden" }`

There is no Vercel Cron config for this — `apps/server` isn't hosted on Vercel. Trigger it with
whatever scheduler sits in front of the deployed server (Railway cron, a scheduled GitHub Action,
etc).

---

## Error Responses

All errors:
```json
{ "error": "message" }
```

| Status | Meaning |
|---|---|
| `400` | Invalid request body |
| `401` | Missing/invalid token, or wrong credentials/TOTP code |
| `403` | Wrong `CRON_SECRET` |
| `404` | Resource not found (or belongs to another user — same response, to avoid leaking existence) |
| `500` | Internal server error |

## CORS Headers

All routes include, when the request's `Origin` is in `ALLOWED_ORIGINS`:
```
Access-Control-Allow-Origin: <matched origin>
Access-Control-Allow-Credentials: true
Access-Control-Allow-Methods: GET, POST, PATCH, DELETE, OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization
Access-Control-Max-Age: 86400
```
`OPTIONS` preflight requests return `204` with these headers.
