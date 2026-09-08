package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// User represents the single authenticated user of the application.
type User struct {
	ID                bson.ObjectID `bson:"_id,omitempty"          json:"id"`
	Email             string        `bson:"email"                  json:"email"`
	PasswordHash      string        `bson:"password_hash"          json:"-"`
	TOTPSecret        string        `bson:"totp_secret"            json:"-"`
	TOTPEnabled       bool          `bson:"totp_enabled"           json:"totp_enabled"`
	TOTPPendingSecret string        `bson:"totp_pending_secret"    json:"-"`
	BackupCodes       []string      `bson:"backup_codes"           json:"-"`
	CreatedAt         time.Time     `bson:"created_at"             json:"created_at"`
}

// PushSubscription holds a Web Push subscription for a user's browser/device.
type PushSubscription struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    bson.ObjectID `bson:"user_id"       json:"-"`
	Endpoint  string        `bson:"endpoint"      json:"endpoint"`
	P256dh    string        `bson:"p256dh"        json:"-"`
	Auth      string        `bson:"auth"          json:"-"`
	CreatedAt time.Time     `bson:"created_at"    json:"created_at"`
}

// Context represents a life area or job (e.g. "Job 1", "Personal", "Health").
type Context struct {
	ID          bson.ObjectID `bson:"_id,omitempty"        json:"id"`
	UserID      bson.ObjectID `bson:"user_id"              json:"-"`
	Name        string        `bson:"name"                 json:"name"`
	Slug        string        `bson:"slug"                 json:"slug"`
	Color       string        `bson:"color"                json:"color"`
	Icon        string        `bson:"icon"                 json:"icon"`
	Type        string        `bson:"type"                 json:"type"` // work | personal | health | finance | travel | custom
	Description string        `bson:"description,omitempty" json:"description,omitempty"`
	Order       int           `bson:"order"                json:"order"`
	CreatedAt   time.Time     `bson:"created_at"           json:"created_at"`
	UpdatedAt   time.Time     `bson:"updated_at"           json:"updated_at"`
}

// Task is the core entity — a single actionable item.
type Task struct {
	ID          bson.ObjectID  `bson:"_id,omitempty"          json:"id"`
	UserID      bson.ObjectID  `bson:"user_id"                json:"-"`
	ContextID   bson.ObjectID  `bson:"context_id"             json:"context_id"`
	Title       string         `bson:"title"                  json:"title"`
	Description string         `bson:"description,omitempty"  json:"description,omitempty"`
	Priority    string         `bson:"priority"               json:"priority"` // low | medium | high | urgent
	Status      string         `bson:"status"                 json:"status"`   // todo | in_progress | done | archived
	DueDate     *time.Time     `bson:"due_date,omitempty"     json:"due_date,omitempty"`
	ReminderAt  *time.Time     `bson:"reminder_at,omitempty"  json:"reminder_at,omitempty"`
	Recurrence  string         `bson:"recurrence,omitempty"   json:"recurrence,omitempty"` // none | daily | weekly | monthly
	Tags        []string       `bson:"tags"                   json:"tags"`
	Notes       string         `bson:"notes,omitempty"        json:"notes,omitempty"`
	CompletedAt *time.Time     `bson:"completed_at,omitempty" json:"completed_at,omitempty"`
	CreatedAt   time.Time      `bson:"created_at"             json:"created_at"`
	UpdatedAt   time.Time      `bson:"updated_at"             json:"updated_at"`
}

// Event is a calendar entry — something that happens at a specific date/time.
type Event struct {
	ID          bson.ObjectID  `bson:"_id,omitempty"         json:"id"`
	UserID      bson.ObjectID  `bson:"user_id"               json:"-"`
	ContextID   *bson.ObjectID `bson:"context_id,omitempty"  json:"context_id,omitempty"`
	Title       string         `bson:"title"                 json:"title"`
	Type        string         `bson:"type"                  json:"type"` // payroll | vacation | deadline | appointment | custom
	Description string         `bson:"description,omitempty" json:"description,omitempty"`
	StartTime   time.Time      `bson:"start_time"            json:"start_time"`
	EndTime     *time.Time     `bson:"end_time,omitempty"    json:"end_time,omitempty"`
	AllDay      bool           `bson:"all_day"               json:"all_day"`
	Recurrence  string         `bson:"recurrence"            json:"recurrence"` // none | monthly | annually
	CreatedAt   time.Time      `bson:"created_at"            json:"created_at"`
	UpdatedAt   time.Time      `bson:"updated_at"            json:"updated_at"`
}

// Credential is an encrypted password vault entry.
// The server stores and returns ciphertext only — never decrypts.
type Credential struct {
	ID            bson.ObjectID `bson:"_id,omitempty"  json:"id"`
	UserID        bson.ObjectID `bson:"user_id"        json:"-"`
	Site          string        `bson:"site"           json:"site"`
	SiteURL       string        `bson:"site_url"       json:"site_url"`
	Username      string        `bson:"username"       json:"username"`
	EncryptedData string        `bson:"encrypted_data" json:"encrypted_data"`
	IV            string        `bson:"iv"             json:"iv"`
	Salt          string        `bson:"salt"           json:"salt"`
	Notes         string        `bson:"notes,omitempty" json:"notes,omitempty"`
	Tags          []string      `bson:"tags"           json:"tags"`
	CreatedAt     time.Time     `bson:"created_at"     json:"created_at"`
	UpdatedAt     time.Time     `bson:"updated_at"     json:"updated_at"`
}

// Notification is an in-app notification record.
type Notification struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    bson.ObjectID `bson:"user_id"       json:"-"`
	Title     string        `bson:"title"         json:"title"`
	Body      string        `bson:"body"          json:"body"`
	Read      bool          `bson:"read"          json:"read"`
	CreatedAt time.Time     `bson:"created_at"    json:"created_at"`
}
