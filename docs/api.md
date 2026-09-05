# API Reference

All API routes are Next.js Route Handlers under `app/api/`.
All routes except `/api/auth/**` require a valid session (Auth.js JWT cookie).
All responses are `Content-Type: application/json`.

---

## Authentication

### `GET|POST /api/auth/[...nextauth]`
Handled entirely by Auth.js. Endpoints:
- `POST /api/auth/signin` — sign in with credentials
- `POST /api/auth/signout` — sign out
- `GET /api/auth/session` — get current session
- `GET /api/auth/csrf` — CSRF token

---

## Tasks

### `GET /api/tasks`
Returns all tasks, optionally filtered.

Query params:
| Param | Type | Description |
|---|---|---|
| `contextId` | string | Filter by context ObjectId |
| `status` | string | `todo`, `in_progress`, `done`, `archived` |
| `priority` | string | `low`, `medium`, `high`, `urgent` |
| `tag` | string | Filter by tag |
| `dueToday` | boolean | Returns tasks with deadline today |
| `overdue` | boolean | Returns tasks past deadline |

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

---

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

Response: `201 Created` with the created task document.

---

### `GET /api/tasks/[id]`
Returns a single task by ObjectId.

Response: `200` with task, or `404` if not found.

---

### `PATCH /api/tasks/[id]`
Partial update of a task. Send only the fields to change.

Request body: any subset of task fields.

Response: `200` with updated task.

---

### `DELETE /api/tasks/[id]`
Deletes a task.

Response: `200 { "deleted": true }`, or `404`.

---

## Contexts

### `GET /api/contexts`
Returns all contexts, sorted by `order`.

### `POST /api/contexts`
Creates a new context.

Request body:
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
Deletes a context. Reassigns orphaned tasks to a default context.

---

## Events

### `GET /api/events`
Returns events, optionally filtered by date range.

Query params:
| Param | Type | Description |
|---|---|---|
| `from` | ISO date | Start of range |
| `to` | ISO date | End of range |
| `type` | string | `payroll`, `vacation`, `deadline`, etc. |

### `POST /api/events`
Creates a new event.

Request body:
```json
{
  "title": "string",
  "type": "payroll | vacation | deadline | appointment | custom",
  "contextId": "ObjectId string (optional)",
  "date": "ISO date string",
  "endDate": "ISO date string (optional)",
  "allDay": true,
  "recurrence": "none | monthly | annually",
  "notes": "string (optional)"
}
```

### `PATCH /api/events/[id]`
Updates an event.

### `DELETE /api/events/[id]`
Deletes an event.

---

## Credentials (Password Vault)

### `GET /api/credentials`
Returns all credentials. **Encrypted fields are returned as-is — no server-side decryption.**

Query params:
| Param | Type | Description |
|---|---|---|
| `site` | string | Filter by siteUrl (for extension domain lookup) |
| `tag` | string | Filter by tag |

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
Stores a new encrypted credential.

Request body:
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

The server never decrypts, modifies, or re-encrypts this data. It stores it verbatim.

### `GET /api/credentials/[id]`
Returns a single credential (still encrypted).

### `PATCH /api/credentials/[id]`
Updates a credential. The client must re-encrypt before sending.

### `DELETE /api/credentials/[id]`
Deletes a credential permanently.

---

## Notifications

### `GET /api/notifications`
Returns notification history.

Query params:
| Param | Type | Description |
|---|---|---|
| `read` | boolean | Filter by read status |
| `limit` | number | Max results (default 50) |

### `PATCH /api/notifications/[id]`
Marks a notification as read.

Request body: `{ "read": true }`

---

## Push Notifications

### `POST /api/push/subscribe`
Saves a browser push subscription for the user.

Request body: the `PushSubscription` object from `navigator.serviceWorker.pushManager.subscribe()`.

### `POST /api/push/send`
Internal route — triggers a push notification manually.
Protected by session (not for external use).

---

## Cron

### `GET /api/cron/notify`
Called by Vercel Cron Job daily at 8:00 AM.

**Protected by `CRON_SECRET` header — not a user-facing endpoint.**

Required header: `Authorization: Bearer <CRON_SECRET>`

Logic:
1. Find tasks with `deadline = today` or overdue
2. Find events within 2 days (payroll, travel, etc.)
3. Send Web Push to all stored subscriptions
4. Log to `notifications` collection

Response: `200 { "sent": 5 }` or `401` if secret is missing/invalid.

---

## Error Responses

All errors follow this shape:

```json
{
  "error": "Human-readable error message"
}
```

| Status | Meaning |
|---|---|
| `400` | Invalid request body (Zod validation failed) |
| `401` | Not authenticated |
| `403` | Authenticated but not authorized (wrong secret) |
| `404` | Resource not found |
| `500` | Internal server error |
