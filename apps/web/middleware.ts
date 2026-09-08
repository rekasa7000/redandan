import { type NextRequest, NextResponse } from "next/server";
import { jwtVerify } from "jose";

import { SESSION_COOKIE } from "@/lib/session";

const secret = () => new TextEncoder().encode(process.env.AUTH_SECRET!);

export async function middleware(req: NextRequest) {
  const { pathname } = req.nextUrl;

  // Redirect authenticated users away from the login page
  if (pathname === "/login") {
    const token = req.cookies.get(SESSION_COOKIE)?.value;
    if (token) {
      try {
        await jwtVerify(token, secret());
        return NextResponse.redirect(new URL("/", req.url));
      } catch {
        // Invalid token — let them through to login
      }
    }
    return NextResponse.next();
  }

  // Protect all other matched routes
  const token = req.cookies.get(SESSION_COOKIE)?.value;
  if (!token) {
    return NextResponse.redirect(new URL("/login", req.url));
  }

  try {
    await jwtVerify(token, secret());
    return NextResponse.next();
  } catch {
    const res = NextResponse.redirect(new URL("/login", req.url));
    res.cookies.delete(SESSION_COOKIE);
    return res;
  }
}

export const config = {
  // Run on all routes except Next.js internals, static files, and API routes
  // (API routes do their own auth via validateCaller)
  matcher: ["/((?!_next/static|_next/image|favicon.ico|api).*)"],
};
