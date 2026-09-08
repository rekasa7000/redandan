import { NextRequest, NextResponse } from "next/server";

import { getDb } from "@/lib/db";
import { validateCaller } from "@/lib/auth";
import { generateSecret, getKeyUri, generateQRDataURL } from "@/lib/totp";
import { ObjectId } from "mongodb";
import type { UserDoc } from "@/lib/types";
// Response shape contract from openapi.yaml
import type { components } from "@/lib/types.gen";

type Res = components["schemas"]["TOTPSetupResponse"];

export async function GET(req: NextRequest) {
  const userId = await validateCaller(req);
  if (!userId) {
    return NextResponse.json({ error: "unauthorized" }, { status: 401 });
  }

  const db = await getDb();
  const user = await db.collection<UserDoc>("users").findOne({
    _id: new ObjectId(userId),
  });
  if (!user) {
    return NextResponse.json({ error: "unauthorized" }, { status: 401 });
  }

  const secret = generateSecret();
  const uri = getKeyUri(user.email, secret);
  const qr = await generateQRDataURL(uri);

  await db.collection<UserDoc>("users").updateOne(
    { _id: user._id },
    { $set: { totpPendingSecret: secret } },
  );

  const res: Res = { uri, secret };
  return NextResponse.json({ ...res, qr });
}
