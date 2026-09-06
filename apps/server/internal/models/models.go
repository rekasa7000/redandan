package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson/primitive"
)

// User represents the single authenticated user of the application.
type User struct {
	ID                primitive.ObjectID `bson:"_id,omitempty"        json:"_id"`
	Username          string             `bson:"username"             json:"username"`
	PasswordHash      string             `bson:"passwordHash"         json:"-"`
	TOTPSecret        string             `bson:"totpSecret"           json:"-"` // AES-GCM encrypted with JWT_SECRET
	TOTPEnabled       bool               `bson:"totpEnabled"          json:"totpEnabled"`
	BackupCodes       []string           `bson:"backupCodes"          json:"-"` // bcrypt-hashed
	PushSubscriptions []PushSubscription `bson:"pushSubscriptions"    json:"-"`
	CreatedAt         time.Time          `bson:"createdAt"            json:"createdAt"`
}

// PushSubscription holds a Web Push subscription object.
type PushSubscription struct {
	Endpoint string            `bson:"endpoint" json:"endpoint"`
	Keys     map[string]string `bson:"keys"     json:"keys"`
}

// Context represents a life area or job (e.g. "Job 1", "Personal", "Health").
type Context struct {
	ID    primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	Name  string             `bson:"name"          json:"name"`
	Slug  string             `bson:"slug"          json:"slug"`
	Color string             `bson:"color"         json:"color"`
	Icon  string             `bson:"icon"          json:"icon"`
	Type  string             `bson:"type"          json:"type"` // work | personal | health | finance | travel | custom
	Order int                `bson:"order"         json:"order"`
}

// Task is the core entity — a single actionable item.
type Task struct {
	ID          primitive.ObjectID  `bson:"_id,omitempty"        json:"_id"`
	Title       string              `bson:"title"                json:"title"`
	Description string              `bson:"description,omitempty" json:"description,omitempty"`
	ContextID   primitive.ObjectID  `bson:"contextId"            json:"contextId"`
	Priority    string              `bson:"priority"             json:"priority"`    // low | medium | high | urgent
	Status      string              `bson:"status"               json:"status"`      // todo | in_progress | done | archived
	Deadline    *time.Time          `bson:"deadline,omitempty"   json:"deadline,omitempty"`
	ReminderAt  *time.Time          `bson:"reminderAt,omitempty" json:"reminderAt,omitempty"`
	Recurrence  string              `bson:"recurrence,omitempty" json:"recurrence,omitempty"` // none | daily | weekly | monthly
	Tags        []string            `bson:"tags"                 json:"tags"`
	Notes       string              `bson:"notes,omitempty"      json:"notes,omitempty"`
	CreatedAt   time.Time           `bson:"createdAt"            json:"createdAt"`
	UpdatedAt   time.Time           `bson:"updatedAt"            json:"updatedAt"`
}

// Event is a calendar entry — something that happens at a specific date/time.
type Event struct {
	ID         primitive.ObjectID  `bson:"_id,omitempty"        json:"_id"`
	Title      string              `bson:"title"                json:"title"`
	Type       string              `bson:"type"                 json:"type"` // payroll | vacation | deadline | appointment | custom
	ContextID  *primitive.ObjectID `bson:"contextId,omitempty"  json:"contextId,omitempty"`
	Date       time.Time           `bson:"date"                 json:"date"`
	EndDate    *time.Time          `bson:"endDate,omitempty"    json:"endDate,omitempty"`
	AllDay     bool                `bson:"allDay"               json:"allDay"`
	Recurrence string              `bson:"recurrence"           json:"recurrence"` // none | monthly | annually
	Notes      string              `bson:"notes,omitempty"      json:"notes,omitempty"`
	CreatedAt  time.Time           `bson:"createdAt"            json:"createdAt"`
}

// Credential is an encrypted password vault entry.
// The server stores and returns ciphertext only — never decrypts.
type Credential struct {
	ID                primitive.ObjectID `bson:"_id,omitempty"           json:"_id"`
	Site              string             `bson:"site"                    json:"site"`
	SiteURL           string             `bson:"siteUrl"                 json:"siteUrl"`
	Username          string             `bson:"username"                json:"username"`
	EncryptedPassword string             `bson:"encryptedPassword"       json:"encryptedPassword"`
	IV                string             `bson:"iv"                      json:"iv"`
	Salt              string             `bson:"salt"                    json:"salt"`
	EncryptedNotes    string             `bson:"encryptedNotes,omitempty" json:"encryptedNotes,omitempty"`
	NotesIV           string             `bson:"notesIv,omitempty"       json:"notesIv,omitempty"`
	Tags              []string           `bson:"tags"                    json:"tags"`
	LastModified      time.Time          `bson:"lastModified"            json:"lastModified"`
	CreatedAt         time.Time          `bson:"createdAt"               json:"createdAt"`
}

// Notification is a record of a sent push notification.
type Notification struct {
	ID      primitive.ObjectID  `bson:"_id,omitempty"      json:"_id"`
	Type    string              `bson:"type"               json:"type"` // task_due | task_overdue | event_reminder | overdue_digest
	RefID   *primitive.ObjectID `bson:"refId,omitempty"    json:"refId,omitempty"`
	RefType string              `bson:"refType,omitempty"  json:"refType,omitempty"` // task | event
	Message string              `bson:"message"            json:"message"`
	SentAt  time.Time           `bson:"sentAt"             json:"sentAt"`
	Read    bool                `bson:"read"               json:"read"`
}
