import { MongoClient, Db } from "mongodb";

const uri = process.env.MONGODB_URI!;
const dbName = process.env.DB_NAME ?? "reliva";

// In development, reuse the connection across hot reloads to avoid
// exhausting the MongoDB Atlas connection pool.
declare global {
  // eslint-disable-next-line no-var
  var _mongoClient: MongoClient | undefined;
}

function getClient(): MongoClient {
  if (process.env.NODE_ENV === "development") {
    if (!global._mongoClient) {
      global._mongoClient = new MongoClient(uri);
    }
    return global._mongoClient;
  }
  return new MongoClient(uri);
}

const client = getClient();

export async function getDb(): Promise<Db> {
  await client.connect();
  return client.db(dbName);
}
