import { SignJWT, jwtVerify } from "jose";
import { cookies } from "next/headers";

export const SESSION_COOKIE = "reliva_session";
export const PENDING_COOKIE = "reliva_pending";

const secret = () => new TextEncoder().encode(process.env.AUTH_SECRET!);

export interface SessionPayload {
  userId: string;
}

// ── Sign / verify ────────────────────────────────────────────

export async function signToken(
  payload: SessionPayload,
  ttl: string,
): Promise<string> {
  return new SignJWT({ ...payload })
    .setProtectedHeader({ alg: "HS256" })
    .setIssuedAt()
    .setExpirationTime(ttl)
    .sign(secret());
}

export async function verifyToken(
  token: string,
): Promise<SessionPayload | null> {
  try {
    const { payload } = await jwtVerify(token, secret());
    return { userId: payload.userId as string };
  } catch {
    return null;
  }
}

// ── Cookie helpers (server components / route handlers only) ─

export async function getSession(): Promise<SessionPayload | null> {
  const store = await cookies();
  const token = store.get(SESSION_COOKIE)?.value;
  if (!token) return null;
  return verifyToken(token);
}

export async function setSession(userId: string): Promise<void> {
  const token = await signToken({ userId }, "30d");
  const store = await cookies();
  store.set(SESSION_COOKIE, token, {
    httpOnly: true,
    secure: process.env.NODE_ENV === "production",
    sameSite: "lax",
    maxAge: 60 * 60 * 24 * 30,
    path: "/",
  });
}

export async function clearSession(): Promise<void> {
  const store = await cookies();
  store.delete(SESSION_COOKIE);
}

export async function setPending(userId: string): Promise<void> {
  const token = await signToken({ userId }, "5m");
  const store = await cookies();
  store.set(PENDING_COOKIE, token, {
    httpOnly: true,
    secure: process.env.NODE_ENV === "production",
    sameSite: "lax",
    maxAge: 60 * 5,
    path: "/",
  });
}

export async function getPending(): Promise<SessionPayload | null> {
  const store = await cookies();
  const token = store.get(PENDING_COOKIE)?.value;
  if (!token) return null;
  return verifyToken(token);
}

export async function clearPending(): Promise<void> {
  const store = await cookies();
  store.delete(PENDING_COOKIE);
}
