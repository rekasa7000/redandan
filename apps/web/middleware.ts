import { type NextRequest, NextResponse } from "next/server";
import { TOKEN_COOKIE } from "@/lib/session";

export function middleware(req: NextRequest) {
  const { pathname } = req.nextUrl;
  const token = req.cookies.get(TOKEN_COOKIE)?.value;

  // Redirect authenticated users away from the login page
  if (pathname === "/login") {
    if (token) return NextResponse.redirect(new URL("/", req.url));
    return NextResponse.next();
  }

  // Protect all matched routes — the Go server validates the JWT on every
  // data request; here we only need presence to avoid an unnecessary redirect.
  if (!token) {
    return NextResponse.redirect(new URL("/login", req.url));
  }

  return NextResponse.next();
}

export const config = {
  matcher: ["/((?!_next/static|_next/image|favicon.ico).*)"],
};
