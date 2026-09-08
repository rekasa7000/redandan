"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import { getToken, clearToken } from "@/lib/session";

const API = process.env.NEXT_PUBLIC_API_URL;

function authHeaders() {
  return {
    "Content-Type": "application/json",
    Authorization: `Bearer ${getToken() ?? ""}`,
  };
}

type TOTPStep = "idle" | "scan" | "done";

export default function SettingsPage() {
  const router = useRouter();

  // User profile
  const [totpEnabled, setTotpEnabled] = useState<boolean | null>(null);

  // TOTP setup
  const [totpStep, setTotpStep] = useState<TOTPStep>("idle");
  const [qr, setQr] = useState("");
  const [secret, setSecret] = useState("");
  const [totpCode, setTotpCode] = useState("");
  const [backupCodes, setBackupCodes] = useState<string[]>([]);
  const [totpError, setTotpError] = useState("");
  const [totpLoading, setTotpLoading] = useState(false);

  // Change password
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [pwError, setPwError] = useState("");
  const [pwSuccess, setPwSuccess] = useState("");
  const [pwLoading, setPwLoading] = useState(false);

  // Logout
  const [logoutLoading, setLogoutLoading] = useState(false);

  useEffect(() => {
    fetch(`${API}/api/v1/auth/me`, { headers: authHeaders() })
      .then((r) => r.json())
      .then((d) => setTotpEnabled(d.totp_enabled ?? false))
      .catch(() => setTotpEnabled(false));
  }, []);

  // ── TOTP setup ────────────────────────────────────────────────

  async function startTOTPSetup() {
    setTotpError("");
    setTotpLoading(true);
    try {
      const res = await fetch(`${API}/api/v1/auth/totp/setup`, {
        headers: authHeaders(),
      });
      const data = await res.json();
      if (!res.ok) {
        setTotpError(data.error ?? "Setup failed");
        return;
      }
      setQr(data.qr);
      setSecret(data.secret);
      setTotpStep("scan");
    } catch {
      setTotpError("Something went wrong");
    } finally {
      setTotpLoading(false);
    }
  }

  async function confirmTOTPSetup(e: React.FormEvent) {
    e.preventDefault();
    setTotpError("");
    setTotpLoading(true);
    try {
      const res = await fetch(`${API}/api/v1/auth/totp/confirm`, {
        method: "POST",
        headers: authHeaders(),
        body: JSON.stringify({ code: totpCode }),
      });
      const data = await res.json();
      if (!res.ok) {
        setTotpError(data.error ?? "Invalid code");
        return;
      }
      setBackupCodes(data.backup_codes);
      setTotpEnabled(true);
      setTotpStep("done");
    } catch {
      setTotpError("Something went wrong");
    } finally {
      setTotpLoading(false);
    }
  }

  // ── Change password ───────────────────────────────────────────

  async function handleChangePassword(e: React.FormEvent) {
    e.preventDefault();
    setPwError("");
    setPwSuccess("");

    if (newPassword !== confirmPassword) {
      setPwError("Passwords do not match");
      return;
    }

    setPwLoading(true);
    try {
      const res = await fetch(`${API}/api/v1/auth/change-password`, {
        method: "POST",
        headers: authHeaders(),
        body: JSON.stringify({
          current_password: currentPassword,
          new_password: newPassword,
        }),
      });
      const data = await res.json();
      if (!res.ok) {
        setPwError(data.error ?? "Failed to change password");
        return;
      }
      setPwSuccess("Password updated.");
      setCurrentPassword("");
      setNewPassword("");
      setConfirmPassword("");
    } catch {
      setPwError("Something went wrong");
    } finally {
      setPwLoading(false);
    }
  }

  // ── Logout ────────────────────────────────────────────────────

  async function handleLogout() {
    setLogoutLoading(true);
    try {
      await fetch(`${API}/api/v1/auth/logout`, {
        method: "POST",
        headers: authHeaders(),
      });
    } finally {
      clearToken();
      router.replace("/login");
    }
  }

  // ── TOTP done — show backup codes ─────────────────────────────

  if (totpStep === "done") {
    return (
      <div className="flex min-h-svh items-center justify-center p-4">
        <Card className="w-full max-w-sm">
          <CardHeader>
            <CardTitle>TOTP enabled</CardTitle>
            <CardDescription>
              Save these backup codes somewhere safe. Each can only be used once.
              You will need them to recover your account if you lose access to
              your authenticator app.
            </CardDescription>
          </CardHeader>
          <CardContent className="flex flex-col gap-4">
            <div className="grid grid-cols-2 gap-2">
              {backupCodes.map((c) => (
                <code
                  key={c}
                  className="rounded bg-muted px-2 py-1 text-center text-xs font-mono"
                >
                  {c}
                </code>
              ))}
            </div>
            <Button onClick={() => router.replace("/")}>Go to dashboard</Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  // ── TOTP scan step ────────────────────────────────────────────

  if (totpStep === "scan") {
    return (
      <div className="flex min-h-svh items-center justify-center p-4">
        <Card className="w-full max-w-sm">
          <CardHeader>
            <CardTitle>Scan QR code</CardTitle>
            <CardDescription>
              Open Google Authenticator (or any TOTP app) and scan the code
              below to add Reliva.
            </CardDescription>
          </CardHeader>
          <CardContent className="flex flex-col gap-4">
            {qr && (
              // eslint-disable-next-line @next/next/no-img-element
              <img src={qr} alt="TOTP QR code" className="mx-auto rounded" />
            )}
            <p className="text-center text-xs text-muted-foreground">
              Can&apos;t scan? Enter this key manually:
            </p>
            <code className="break-all rounded bg-muted px-2 py-1 text-center text-xs font-mono">
              {secret}
            </code>
            <form onSubmit={confirmTOTPSetup} className="flex flex-col gap-4">
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="totp-code">Confirm with a code from your app</Label>
                <Input
                  id="totp-code"
                  type="text"
                  inputMode="numeric"
                  maxLength={6}
                  value={totpCode}
                  onChange={(e) => setTotpCode(e.target.value)}
                  placeholder="000000"
                  required
                />
              </div>
              {totpError && <p className="text-sm text-destructive">{totpError}</p>}
              <Button type="submit" disabled={totpLoading}>
                {totpLoading ? "Verifying…" : "Enable TOTP"}
              </Button>
            </form>
          </CardContent>
        </Card>
      </div>
    );
  }

  // ── Main settings page ────────────────────────────────────────

  return (
    <div className="mx-auto max-w-lg space-y-6 p-6">
      <h1 className="text-2xl font-semibold tracking-tight">Settings</h1>

      {/* Two-factor authentication */}
      <Card>
        <CardHeader>
          <CardTitle>Two-factor authentication</CardTitle>
          <CardDescription>
            {totpEnabled === null
              ? "Loading…"
              : totpEnabled
                ? "TOTP is active. You can reset it to re-link your authenticator app."
                : "TOTP is required. Set it up to access Reliva."}
          </CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col gap-3">
          {totpError && <p className="text-sm text-destructive">{totpError}</p>}
          <Button
            onClick={startTOTPSetup}
            disabled={totpLoading || totpEnabled === null}
            variant={totpEnabled ? "outline" : "default"}
          >
            {totpLoading
              ? "Loading…"
              : totpEnabled
                ? "Reset authenticator"
                : "Set up authenticator"}
          </Button>
        </CardContent>
      </Card>

      {/* Change password */}
      <Card>
        <CardHeader>
          <CardTitle>Change password</CardTitle>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleChangePassword} className="flex flex-col gap-4">
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="current-password">Current password</Label>
              <Input
                id="current-password"
                type="password"
                value={currentPassword}
                onChange={(e) => setCurrentPassword(e.target.value)}
                required
              />
            </div>
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="new-password">New password</Label>
              <Input
                id="new-password"
                type="password"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                minLength={8}
                required
              />
            </div>
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="confirm-password">Confirm new password</Label>
              <Input
                id="confirm-password"
                type="password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                minLength={8}
                required
              />
            </div>
            {pwError && <p className="text-sm text-destructive">{pwError}</p>}
            {pwSuccess && <p className="text-sm text-green-600">{pwSuccess}</p>}
            <Button type="submit" disabled={pwLoading}>
              {pwLoading ? "Saving…" : "Update password"}
            </Button>
          </form>
        </CardContent>
      </Card>

      {/* Logout */}
      <Card>
        <CardHeader>
          <CardTitle>Session</CardTitle>
        </CardHeader>
        <CardContent>
          <Button
            variant="destructive"
            onClick={handleLogout}
            disabled={logoutLoading}
            className="w-full"
          >
            {logoutLoading ? "Signing out…" : "Sign out"}
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}
