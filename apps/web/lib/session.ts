// Cookie helpers for the Go server's access JWT.
// The token is issued by apps/server (Go + Gin) and stored in a browser
// cookie. The Next.js app never signs or verifies JWTs itself.

export const TOKEN_COOKIE = "reliva_token";

/** Read the Go server token from a cookie string (edge-compatible). */
export function getTokenFromCookieHeader(cookieHeader: string): string | null {
  const match = cookieHeader.match(
    new RegExp(`(?:^|;\\s*)${TOKEN_COOKIE}=([^;]+)`),
  );
  return match ? decodeURIComponent(match[1]) : null;
}

/** Set the reliva_token cookie (call from client-side code after login). */
export function setToken(token: string, maxAgeDays = 30): void {
  const maxAge = maxAgeDays * 24 * 60 * 60;
  const secure = location.protocol === "https:" ? "; Secure" : "";
  document.cookie = `${TOKEN_COOKIE}=${encodeURIComponent(token)}; Max-Age=${maxAge}; Path=/${secure}; SameSite=Lax`;
}

/** Clear the reliva_token cookie (call from client-side code on logout). */
export function clearToken(): void {
  document.cookie = `${TOKEN_COOKIE}=; Max-Age=0; Path=/`;
}

/** Read the token from document.cookie (client-side only). */
export function getToken(): string | null {
  if (typeof document === "undefined") return null;
  return getTokenFromCookieHeader(document.cookie);
}
