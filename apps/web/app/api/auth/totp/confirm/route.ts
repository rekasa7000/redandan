import { type NextRequest, NextResponse } from "next/server";
import { z } from "zod";
import { hash } from "bcryptjs";
import { randomBytes } from "crypto";

import { getDb } from "@/lib/db";
import { validateCaller } from "@/lib/auth";
import { verifyCode } from "@/lib/totp";
import { ObjectId } from "mongodb";
import type { UserDoc } from "@/lib/types";
// Response shape contract from openapi.yaml
import type { components } from "@/lib/types.gen";

type Res = components["schemas"]["TOTPConfirmResponse"];

const schema = z.object({ code: z.string().length(6) });

export async function POST(req: NextRequest) {
  const userId = await validateCaller(req);
  if (!userId) {
    return NextResponse.json({ error: "unauthorized" }, { status: 401 });
  }

  const body = await req.json().catch(() => null);
  const parsed = schema.safeParse(body);
  if (!parsed.success) {
    return NextResponse.json({ error: "code required" }, { status: 400 });
  }

  const db = await getDb();
  const user = await db.collection<UserDoc>("users").findOne({
    _id: new ObjectId(userId),
  });
  if (!user) {
    return NextResponse.json({ error: "unauthorized" }, { status: 401 });
  }
  if (!user.totpPendingSecret) {
    return NextResponse.json(
      { error: "no pending TOTP setup; call /totp/setup first" },
      { status: 400 },
    );
  }

  if (!verifyCode(parsed.data.code, user.totpPendingSecret)) {
    return NextResponse.json({ error: "invalid TOTP code" }, { status: 401 });
  }

  // Generate 8 backup codes — plaintext returned once, stored as bcrypt hashes.
  const plainCodes = Array.from({ length: 8 }, () =>
    randomBytes(6).toString("hex"),
  );
  const hashedCodes = await Promise.all(plainCodes.map((c) => hash(c, 10)));

  await db.collection<UserDoc>("users").updateOne(
    { _id: user._id },
    {
      $set: {
        totpSecret: user.totpPendingSecret,
        totpEnabled: true,
        totpPendingSecret: "",
        backupCodes: hashedCodes,
      },
    },
  );

  const res: Res = { message: "TOTP enabled", backup_codes: plainCodes };
  return NextResponse.json(res);
}
