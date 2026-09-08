"use client";

import { useState } from "react";
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
import { getToken } from "@/lib/session";

const API = process.env.NEXT_PUBLIC_API_URL;

function authHeaders() {
  return {
    "Content-Type": "application/json",
    Authorization: `Bearer ${getToken() ?? ""}`,
  };
}

type SetupStep = "idle" | "scan" | "done";

export default function SettingsPage() {
  const router = useRouter();
  const [step, setStep] = useState<SetupStep>("idle");
  const [qr, setQr] = useState("");
  const [secret, setSecret] = useState("");
  const [code, setCode] = useState("");
  const [backupCodes, setBackupCodes] = useState<string[]>([]);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  async function startSetup() {
    setError("");
    setLoading(true);
    try {
      const res = await fetch(`${API}/api/v1/auth/totp/setup`, {
        headers: authHeaders(),
      });
      const data = await res.json();
      if (!res.ok) {
        setError(data.error ?? "Setup failed");
        return;
      }
      setQr(data.qr);
      setSecret(data.secret);
      setStep("scan");
    } catch {
      setError("Something went wrong");
    } finally {
      setLoading(false);
    }
  }

  async function confirmSetup(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setLoading(true);
    try {
      const res = await fetch(`${API}/api/v1/auth/totp/confirm`, {
        method: "POST",
        headers: authHeaders(),
        body: JSON.stringify({ code }),
      });
      const data = await res.json();
      if (!res.ok) {
        setError(data.error ?? "Invalid code");
        return;
      }
      setBackupCodes(data.backup_codes);
      setStep("done");
    } catch {
      setError("Something went wrong");
    } finally {
      setLoading(false);
    }
  }

  if (step === "done") {
    return (
      <div className="flex min-h-svh items-center justify-center p-4">
        <Card className="w-full max-w-sm">
          <CardHeader>
            <CardTitle>TOTP enabled</CardTitle>
            <CardDescription>
              Save these backup codes somewhere safe. Each can only be used once.
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

  if (step === "scan") {
    return (
      <div className="flex min-h-svh items-center justify-center p-4">
        <Card className="w-full max-w-sm">
          <CardHeader>
            <CardTitle>Scan QR code</CardTitle>
            <CardDescription>
              Open your authenticator app and scan the QR code below.
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
            <form onSubmit={confirmSetup} className="flex flex-col gap-4">
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="code">Confirm with a code</Label>
                <Input
                  id="code"
                  type="text"
                  inputMode="numeric"
                  maxLength={6}
                  value={code}
                  onChange={(e) => setCode(e.target.value)}
                  placeholder="000000"
                  required
                />
              </div>
              {error && <p className="text-sm text-destructive">{error}</p>}
              <Button type="submit" disabled={loading}>
                {loading ? "Verifying…" : "Enable TOTP"}
              </Button>
            </form>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="flex min-h-svh items-center justify-center p-4">
      <Card className="w-full max-w-sm">
        <CardHeader>
          <CardTitle>Set up two-factor auth</CardTitle>
          <CardDescription>
            TOTP is required before you can access Reliva. Set it up now.
          </CardDescription>
        </CardHeader>
        <CardContent>
          {error && <p className="mb-4 text-sm text-destructive">{error}</p>}
          <Button onClick={startSetup} disabled={loading} className="w-full">
            {loading ? "Loading…" : "Set up authenticator"}
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}
