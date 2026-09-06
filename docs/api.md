# API Reference

All API routes are Next.js Route Handlers under `app/api/`.
All responses are `Content-Type: application/json`.

## Authentication

Every protected route accepts **either** of the following. Both are equally valid:

| Method | Sent by | Header / Cookie |
|---|---|---|
| Session cookie | Web browser (automatic) | `Cookie: next-auth.session-token=...` |
| Bearer token | Extension, Mobile | `Authorization: Bearer <jwt>` |

The `validateCaller()` helper in `lib/auth.ts` handles both methods transparently.
A route that does not call `validateCaller()` is a security bug.

---

## Auth Endpoints

### `POST /api/auth/login`
Step 1 of login. Validates username + password.

Request body:
```json
{ "username": "string", "password": "string" }
```

Response on success: `200 { "totpRequired": true }`
Response on failure: `401 { "error": "Invalid credentials" }`

Does **not** issue a session yet. TOTP must be completed first.

---

### `POST /api/auth/totp/validate`
Step 2 of login. Validates the TOTP code and issues the session or token.

Request body:
```json
{
  "code": "123456",
  "clientType": "web | extension | mobile"
}
```

Response for `web`:
```json
{ "ok": true }
```
Plus an HttpOnly session cookie is set on the response.

Response for `extension` / `mobile`:
```json
{ "token": "<signed JWT>" }
```

Response on failure: `401 { "error": "Invalid code" }`

Rate limited: 5 failed attempts per 15 minutes per IP.

---

### `POST /api/auth/logout`
Invalidates the session cookie (web). For extension/mobile, the client discards the token locally.

Response: `200 { "ok": true }`

---

### `GET /api/auth/totp/setup`
Generates a new TOTP secret and returns the QR code data.
**Requires active session** — called from the settings page.

Response:
```json
{
  "qrCodeDataUrl": "data:image/png;base64,...",
  "secret": "BASE32SECRET",
  "backupCodes": ["XXXXX-XXXXX", "..."]
}
```

The secret and backup code hashes are stored in the `users` collection.
Backup codes are shown only once — the user must save them.

---

### `POST /api/auth/totp/confirm`
Confirms TOTP setup by verifying the user scanned the QR code correctly.
Must be called after `GET /api/auth/totp/setup` to activate TOTP.

Request body:
```json
{ "code": "123456" }
```

Response: `200 { "confirmed": true }` or `400 { "error": "Invalid code" }`

---

### `POST /api/auth/backup-code`
Uses a backup code in place of a TOTP code during login step 2.

Request body:
```json
{ "code": "XXXXX-XXXXX" }
```

Response: same as `POST /api/auth/totp/validate` — issues session or token.
The used backup code is immediately invalidated.

---

### `GET|POST /api/auth/[...nextauth]`
Handled by Auth.js for internal session management. Not called directly by clients.

---

## Tasks

### `GET /api/tasks`
Returns tasks, optionally filtered.

Query params:
| Param | Type | Description |
|---|---|---|
| `contextId` | string | Filter by context ObjectId |
| `status` | string | `todo`, `in_progress`, `done`, `archived` |
| `priority` | string | `low`, `medium`, `high`, `urgent` |
| `tag` | string | Filter by tag |
| `dueToday` | boolean | Tasks with deadline = today |
| `overdue` | boolean | Tasks past deadline, not done |

Response:
```json
{
  "tasks": [
    {
      "_id": "...",
      "title": "Finish report",
      "contextId": "...",
      "priority": "high",
      "status": "todo",
      "deadline": "2026-09-10T00:00:00.000Z",
      "tags": ["report"],
      "createdAt": "...",
      "updatedAt": "..."
    }
  ]
}
```

### `POST /api/tasks`
Creates a new task.

Request body:
```json
{
  "title": "string (required)",
  "description": "string (optional)",
  "contextId": "ObjectId string (required)",
  "priority": "low | medium | high | urgent",
  "status": "todo | in_progress | done | archived",
  "deadline": "ISO date string (optional)",
  "reminderAt": "ISO date string (optional)",
  "recurrence": "none | daily | weekly | monthly",
  "tags": ["string"],
  "notes": "string (optional)"
}
```

Response: `201` with the created task document.

### `GET /api/tasks/[id]`
Returns a single task. `404` if not found.

### `PATCH /api/tasks/[id]`
Partial update. Send only the fields to change.
Response: `200` with the updated task.

### `DELETE /api/tasks/[id]`
Deletes a task. Response: `200 { "deleted": true }`.

---

## Contexts

### `GET /api/contexts`
Returns all contexts, sorted by `order`.

### `POST /api/contexts`
```json
{
  "name": "string",
  "slug": "string",
  "color": "#hexcode",
  "icon": "lucide icon name",
  "type": "work | personal | health | finance | travel | custom",
  "order": 0
}
```

### `PATCH /api/contexts/[id]`
Updates a context.

### `DELETE /api/contexts/[id]`
Deletes a context. Orphaned tasks are reassigned to a default "General" context.

---

## Events

### `GET /api/events`
Query params: `from` (ISO date), `to` (ISO date), `type` (string).

### `POST /api/events`
```json
{
  "title": "string",
  "type": "payroll | vacation | deadline | appointment | custom",
  "contextId": "ObjectId (optional)",
  "date": "ISO date string",
  "endDate": "ISO date string (optional)",
  "allDay": true,
  "recurrence": "none | monthly | annually",
  "notes": "string (optional)"
}
```

### `PATCH /api/events/[id]`
### `DELETE /api/events/[id]`

---

## Credentials (Password Vault)

**The server never decrypts credential data. All encrypted fields are stored and returned verbatim.**

### `GET /api/credentials`
Query params: `site` (string, for domain matching), `tag` (string).

Response:
```json
{
  "credentials": [
    {
      "_id": "...",
      "site": "GitHub",
      "siteUrl": "https://github.com",
      "username": "myusername",
      "encryptedPassword": "base64...",
      "iv": "base64...",
      "salt": "base64...",
      "tags": [],
      "lastModified": "..."
    }
  ]
}
```

### `POST /api/credentials`
```json
{
  "site": "string",
  "siteUrl": "string",
  "username": "string",
  "encryptedPassword": "base64 ciphertext",
  "iv": "base64",
  "salt": "base64",
  "encryptedNotes": "base64 (optional)",
  "notesIv": "base64 (optional)",
  "tags": ["string"]
}
```

### `GET /api/credentials/[id]`
Returns a single credential (ciphertext only).

### `PATCH /api/credentials/[id]`
Updates a credential. The client must re-encrypt before sending updated fields.

### `DELETE /api/credentials/[id]`

---

## Notifications

### `GET /api/notifications`
Query params: `read` (boolean), `limit` (number, default 50).

### `PATCH /api/notifications/[id]`
Request body: `{ "read": true }`

---

## Push Notifications

### `POST /api/push/subscribe`
Saves a browser push subscription.
Request body: `PushSubscription` object from `navigator.serviceWorker.pushManager.subscribe()`.

### `POST /api/push/send`
Internal — triggers a push notification manually. Protected by session.

---

## Cron

### `GET /api/cron/notify`
Called by Vercel Cron Job daily at 8:00 AM.
**Not a user-facing endpoint.**

Required header: `Authorization: Bearer <CRON_SECRET>`

Logic:
1. Find tasks due today or overdue
2. Find events within 2 days
3. Send Web Push to all stored subscriptions
4. Log sent notifications to the `notifications` collection

Response: `200 { "sent": 5 }` or `403` if the secret is wrong.

---

## Error Responses

All errors:
```json
{ "error": "Human-readable message" }
```

| Status | Meaning |
|---|---|
| `400` | Invalid request body (Zod validation failed) |
| `401` | Not authenticated or TOTP failed |
| `403` | Authenticated but forbidden (wrong secret, consumed backup code, etc.) |
| `404` | Resource not found |
| `429` | Rate limited (too many failed auth attempts) |
| `500` | Internal server error |

---

## CORS Headers

All `/api/**` routes include:
```
Access-Control-Allow-Origin: <value from ALLOWED_ORIGINS env var>
Access-Control-Allow-Methods: GET, POST, PATCH, DELETE, OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization
Access-Control-Max-Age: 86400
```

Preflight `OPTIONS` requests return `204` with these headers.
