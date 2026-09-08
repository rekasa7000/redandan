import { type NextRequest, NextResponse } from "next/server";
import { z } from "zod";

import { getDb } from "@/lib/db";
import { getPending, clearPending, setSession } from "@/lib/session";
import { verifyCode } from "@/lib/totp";
import { ObjectId } from "mongodb";
import type { UserDoc } from "@/lib/types";

const schema = z.object({ code: z.string().length(6) });

export async function POST(req: NextRequest) {
  const pending = await getPending();
  if (!pending) {
    return NextResponse.json({ error: "pending token required" }, { status: 401 });
  }

  const body = await req.json().catch(() => null);
  const parsed = schema.safeParse(body);
  if (!parsed.success) {
    return NextResponse.json({ error: "code required" }, { status: 400 });
  }

  const db = await getDb();
  const user = await db.collection<UserDoc>("users").findOne({
    _id: new ObjectId(pending.userId),
  });
  if (!user || !user.totpEnabled || !user.totpSecret) {
    return NextResponse.json({ error: "unauthorized" }, { status: 401 });
  }

  if (!verifyCode(parsed.data.code, user.totpSecret)) {
    return NextResponse.json({ error: "invalid TOTP code" }, { status: 401 });
  }

  await clearPending();
  await setSession(user._id.toString());

  return NextResponse.json({ ok: true });
}
