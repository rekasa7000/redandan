// One-time seed script — run with Node.js (avoids Bun/mongodb compatibility issues).
// Usage: node scripts/seed.cjs --email you@example.com --password yourpassword
//
// Reads MONGO_URI + DB_NAME from apps/server/.env automatically.

const fs = require("fs");
const path = require("path");
const { MongoClient } = require("./node_modules/mongodb");
const { hashSync } = require("./node_modules/bcryptjs");

// Parse apps/server/.env
const envPath = path.join(__dirname, "..", "apps", "server", ".env");
if (fs.existsSync(envPath)) {
  for (const line of fs.readFileSync(envPath, "utf8").split("\n")) {
    const t = line.trim();
    if (!t || t.startsWith("#")) continue;
    const eq = t.indexOf("=");
    if (eq === -1) continue;
    const key = t.slice(0, eq).trim();
    const val = t.slice(eq + 1).trim();
    if (!process.env[key]) process.env[key] = val;
  }
}

// Parse CLI args --email and --password
const args = process.argv.slice(2);
const get = (flag) => {
  const i = args.indexOf(flag);
  return i !== -1 ? args[i + 1] : undefined;
};

const email    = get("--email")    ?? process.env.SEED_EMAIL;
const password = get("--password") ?? process.env.SEED_PASSWORD;
const uri      = process.env.MONGO_URI;
const dbName   = process.env.DB_NAME ?? "reliva";

if (!uri || !email || !password) {
  console.error("Usage: node scripts/seed.cjs --email <email> --password <password>");
  console.error("Ensure MONGO_URI is set in apps/server/.env");
  process.exit(1);
}

async function seed() {
  const client = new MongoClient(uri);
  try {
    await client.connect();
    const db = client.db(dbName);
    const users = db.collection("users");

    const existing = await users.findOne({});
    if (existing) {
      console.log(`User already exists (${existing.email}) — skipping.`);
      console.log("Drop the users collection first if you want to reseed.");
      return;
    }

    const passwordHash = hashSync(password, 12);
    await users.insertOne({
      email,
      passwordHash,
      totpSecret: "",
      totpEnabled: false,
      totpPendingSecret: "",
      backupCodes: [],
      createdAt: new Date(),
    });

    await db.collection("contexts").insertMany([
      { name: "Personal", slug: "personal", color: "#6366f1", icon: "user",      type: "personal", order: 0, createdAt: new Date(), updatedAt: new Date() },
      { name: "Work",     slug: "work",     color: "#f59e0b", icon: "briefcase", type: "work",     order: 1, createdAt: new Date(), updatedAt: new Date() },
      { name: "Health",   slug: "health",   color: "#10b981", icon: "heart",     type: "health",   order: 2, createdAt: new Date(), updatedAt: new Date() },
    ]);

    console.log(`✓ User created: ${email}`);
    console.log("✓ Default contexts: Personal, Work, Health");
    console.log("\nNext: start the server, log in, then set up TOTP in /settings.");
  } finally {
    await client.close();
  }
}

seed().catch((e) => {
  console.error("Seed failed:", e.message);
  process.exit(1);
});
