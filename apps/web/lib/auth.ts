import { type NextRequest } from "next/server";
import { getSession, verifyToken } from "@/lib/session";

// validateCaller returns the authenticated userId from either:
//   1. reliva_session cookie  (web browser)
//   2. Authorization: Bearer  (extension / mobile via bearer token)
// Returns null if neither is valid.
export async function validateCaller(req: NextRequest): Promise<string | null> {
  // Cookie path (server components / route handlers in Node runtime)
  const session = await getSession();
  if (session?.userId) return session.userId;

  // Bearer token path
  const auth = req.headers.get("authorization");
  if (auth?.startsWith("Bearer ")) {
    const token = auth.slice(7);
    const payload = await verifyToken(token);
    return payload?.userId ?? null;
  }

  return null;
}
