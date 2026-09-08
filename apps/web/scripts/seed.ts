// Seed script — run once to create the single admin user.
// Usage: bun run scripts/seed.ts
//
// Required env vars (add to apps/web/.env.local):
//   MONGODB_URI, DB_NAME, SEED_EMAIL, SEED_PASSWORD

import { MongoClient } from "mongodb";
import { hash } from "bcryptjs";

const uri = process.env.MONGODB_URI;
const dbName = process.env.DB_NAME ?? "reliva";
const email = process.env.SEED_EMAIL;
const password = process.env.SEED_PASSWORD;

if (!uri || !email || !password) {
  console.error(
    "Missing required env vars: MONGODB_URI, SEED_EMAIL, SEED_PASSWORD",
  );
  process.exit(1);
}

const client = new MongoClient(uri);

try {
  await client.connect();
  const db = client.db(dbName);
  const users = db.collection("users");

  const existing = await users.findOne({});
  if (existing) {
    console.log("User already exists — skipping seed.");
    process.exit(0);
  }

  const passwordHash = await hash(password, 12);

  await users.insertOne({
    email,
    passwordHash,
    totpSecret: "",
    totpEnabled: false,
    totpPendingSecret: "",
    backupCodes: [],
    createdAt: new Date(),
  });

  // Default contexts
  await db.collection("contexts").insertMany([
    { name: "Personal", slug: "personal", color: "#6366f1", icon: "user", type: "personal", order: 0, createdAt: new Date(), updatedAt: new Date() },
    { name: "Work", slug: "work", color: "#f59e0b", icon: "briefcase", type: "work", order: 1, createdAt: new Date(), updatedAt: new Date() },
    { name: "Health", slug: "health", color: "#10b981", icon: "heart", type: "health", order: 2, createdAt: new Date(), updatedAt: new Date() },
  ]);

  console.log(`✓ User seeded: ${email}`);
  console.log("✓ Default contexts created: Personal, Work, Health");
  console.log("\nNext: open the app, log in, then set up TOTP in /settings.");
} finally {
  await client.close();
}
