// MongoDB document shapes — used server-side only.
// For HTTP request/response shapes, use types.gen.ts (generated from openapi.yaml).

import type { ObjectId } from "mongodb";

export interface UserDoc {
  _id: ObjectId;
  email: string;
  passwordHash: string;
  totpSecret: string;          // active TOTP secret
  totpEnabled: boolean;
  totpPendingSecret: string;   // temporary secret during TOTP setup
  backupCodes: string[];       // bcrypt-hashed one-time codes
  createdAt: Date;
}

export interface ContextDoc {
  _id: ObjectId;
  userId: ObjectId;
  name: string;
  slug: string;
  color: string;
  icon: string;
  type: "work" | "personal" | "health" | "finance" | "travel" | "custom";
  order: number;
  createdAt: Date;
  updatedAt: Date;
}

export interface TaskDoc {
  _id: ObjectId;
  userId: ObjectId;
  contextId?: ObjectId;
  title: string;
  description?: string;
  priority: "low" | "medium" | "high" | "urgent";
  status: "todo" | "in_progress" | "done" | "archived";
  dueDate?: Date;
  reminderAt?: Date;
  recurrence?: "none" | "daily" | "weekly" | "monthly";
  tags: string[];
  notes?: string;
  completedAt?: Date;
  createdAt: Date;
  updatedAt: Date;
}

export interface EventDoc {
  _id: ObjectId;
  userId: ObjectId;
  contextId?: ObjectId;
  title: string;
  description?: string;
  startTime: Date;
  endTime?: Date;
  allDay: boolean;
  recurrence?: "none" | "monthly" | "annually";
  notes?: string;
  createdAt: Date;
  updatedAt: Date;
}

export interface CredentialDoc {
  _id: ObjectId;
  userId: ObjectId;
  site: string;
  siteUrl?: string;
  username: string;
  encryptedData: string;  // AES-GCM ciphertext, base64url
  iv: string;             // AES-GCM IV, base64url
  salt: string;           // PBKDF2 salt, base64url
  encryptedNotes?: string;
  notesIv?: string;
  tags: string[];
  createdAt: Date;
  updatedAt: Date;
}

export interface NotificationDoc {
  _id: ObjectId;
  userId: ObjectId;
  title: string;
  body: string;
  refId?: ObjectId;
  refType?: "task" | "event";
  read: boolean;
  createdAt: Date;
}

export interface PushSubscriptionDoc {
  _id: ObjectId;
  userId: ObjectId;
  endpoint: string;
  p256dh: string;
  auth: string;
  createdAt: Date;
}
