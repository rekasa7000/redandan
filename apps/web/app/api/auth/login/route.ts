import { type NextRequest, NextResponse } from "next/server";
import { compare } from "bcryptjs";
import { z } from "zod";

import { getDb } from "@/lib/db";
import { setSession, setPending } from "@/lib/session";
import type { UserDoc } from "@/lib/types";
// Response shapes contract from openapi.yaml
import type { components } from "@/lib/types.gen";

type PendingRes = components["schemas"]["PendingTokenResponse"];

const schema = z.object({ password: z.string().min(1) });

export async function POST(req: NextRequest) {
  const body = await req.json().catch(() => null);
  const parsed = schema.safeParse(body);
  if (!parsed.success) {
    return NextResponse.json({ error: "password required" }, { status: 400 });
  }

  const db = await getDb();
  const user = await db.collection<UserDoc>("users").findOne({});
  if (!user) {
    return NextResponse.json({ error: "invalid credentials" }, { status: 401 });
  }

  const valid = await compare(parsed.data.password, user.passwordHash);
  if (!valid) {
    return NextResponse.json({ error: "invalid credentials" }, { status: 401 });
  }

  // TOTP not set up yet — issue a full session immediately.
  // The (app) layout will redirect to /settings to complete TOTP setup.
  if (!user.totpEnabled) {
    await setSession(user._id.toString());
    return NextResponse.json({ requireTotp: false });
  }

  // TOTP enabled — issue a short-lived pending cookie for step 2.
  await setPending(user._id.toString());
  const res: PendingRes = { pending_token: "cookie" }; // token is in the cookie, not the body
  return NextResponse.json({ ...res, requireTotp: true });
}
