import { headers } from "next/headers";
import { redirect } from "next/navigation";
import { getSession } from "@/lib/session";
import { getDb } from "@/lib/db";
import { ObjectId } from "mongodb";
import type { UserDoc } from "@/lib/types";

export default async function AppLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const session = await getSession();
  if (!session) redirect("/login");

  const db = await getDb();
  const user = await db.collection<UserDoc>("users").findOne({
    _id: new ObjectId(session.userId),
  });
  if (!user) redirect("/login");

  // First-time setup: redirect to /settings until TOTP is configured.
  // Allow /settings itself to avoid an infinite redirect loop.
  if (!user.totpEnabled) {
    const headersList = await headers();
    const pathname = headersList.get("x-invoke-path") ?? "";
    if (!pathname.startsWith("/settings")) {
      redirect("/settings");
    }
  }

  return <>{children}</>;
}
