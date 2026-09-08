import { type NextRequest, NextResponse } from "next/server";
import { compare } from "bcryptjs";
import { z } from "zod";

import { getDb } from "@/lib/db";
import { getPending, clearPending, setSession } from "@/lib/session";
import { ObjectId } from "mongodb";
import type { UserDoc } from "@/lib/types";

const schema = z.object({ code: z.string().min(1) });

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
  if (!user) {
    return NextResponse.json({ error: "unauthorized" }, { status: 401 });
  }

  // Find matching backup code and burn it.
  const results = await Promise.all(
    user.backupCodes.map((h) => compare(parsed.data.code, h)),
  );
  const matchIdx = results.findIndex(Boolean);
  if (matchIdx === -1) {
    return NextResponse.json({ error: "invalid backup code" }, { status: 401 });
  }

  const remaining = user.backupCodes.filter((_, i) => i !== matchIdx);
  await db.collection<UserDoc>("users").updateOne(
    { _id: user._id },
    { $set: { backupCodes: remaining } },
  );

  await clearPending();
  await setSession(user._id.toString());

  return NextResponse.json({ ok: true });
}
