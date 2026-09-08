// One-time seed script — creates the single Reliva user in MongoDB.
// Usage: bun run scripts/seed.ts --email you@example.com --password yourpassword
//
// Reads MONGO_URI and DB_NAME from apps/server/.env (or env vars).

import { MongoClient } from "mongodb";
import { hash } from "bcryptjs";
import { parseArgs } from "util";

// Load apps/server/.env
const envFile = Bun.file("apps/server/.env");
if (await envFile.exists()) {
  const text = await envFile.text();
  for (const line of text.split("\n")) {
    const trimmed = line.trim();
    if (!trimmed || trimmed.startsWith("#")) continue;
    const eq = trimmed.indexOf("=");
    if (eq === -1) continue;
    const key = trimmed.slice(0, eq).trim();
    const val = trimmed.slice(eq + 1).trim();
    if (!process.env[key]) process.env[key] = val;
  }
}

const { values } = parseArgs({
  args: Bun.argv.slice(2),
  options: {
    email:    { type: "string" },
    password: { type: "string" },
  },
});

const email    = values.email    ?? process.env.SEED_EMAIL;
const password = values.password ?? process.env.SEED_PASSWORD;
const uri      = process.env.MONGO_URI;
const dbName   = process.env.DB_NAME ?? "reliva";

if (!uri || !email || !password) {
  console.error("Usage: bun run scripts/seed.ts --email <email> --password <password>");
  console.error("Also ensure MONGO_URI is set in apps/server/.env");
  process.exit(1);
}

const client = new MongoClient(uri);

try {
  await client.connect();
  const db = client.db(dbName);
  const users = db.collection("users");

  const existing = await users.findOne({});
  if (existing) {
    console.log(`User already exists (${existing.email}) — skipping.`);
    console.log("To reseed, drop the users collection first.");
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
    { name: "Personal", slug: "personal", color: "#6366f1", icon: "user",      type: "personal", order: 0, createdAt: new Date(), updatedAt: new Date() },
    { name: "Work",     slug: "work",     color: "#f59e0b", icon: "briefcase", type: "work",     order: 1, createdAt: new Date(), updatedAt: new Date() },
    { name: "Health",   slug: "health",   color: "#10b981", icon: "heart",     type: "health",   order: 2, createdAt: new Date(), updatedAt: new Date() },
  ]);

  console.log(`✓ User created: ${email}`);
  console.log("✓ Default contexts created: Personal, Work, Health");
  console.log("\nNext: start the server, log in, then set up TOTP in /settings.");
} finally {
  await client.close();
}
