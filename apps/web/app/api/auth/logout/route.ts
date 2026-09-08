import { NextResponse } from "next/server";
import { clearSession, clearPending } from "@/lib/session";
// Response shape contract from openapi.yaml
import type { components } from "@/lib/types.gen";

type Res = components["schemas"]["MessageResponse"];

export async function POST() {
  await clearSession();
  await clearPending();
  const res: Res = { message: "logged out" };
  return NextResponse.json(res);
}
