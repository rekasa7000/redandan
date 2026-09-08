"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";

type Step = "password" | "totp";

export default function LoginPage() {
  const router = useRouter();
  const [step, setStep] = useState<Step>("password");
  const [password, setPassword] = useState("");
  const [code, setCode] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const [useBackup, setUseBackup] = useState(false);

  async function handlePassword(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setLoading(true);

    try {
      const res = await fetch("/api/auth/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ password }),
      });
      const data = await res.json();

      if (!res.ok) {
        setError(data.error ?? "Invalid password");
        return;
      }

      if (data.requireTotp) {
        setStep("totp");
      } else {
        router.replace("/");
      }
    } catch {
      setError("Something went wrong");
    } finally {
      setLoading(false);
    }
  }

  async function handleTotp(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setLoading(true);

    const endpoint = useBackup
      ? "/api/auth/backup-code"
      : "/api/auth/totp/validate";

    try {
      const res = await fetch(endpoint, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ code }),
      });
      const data = await res.json();

      if (!res.ok) {
        setError(data.error ?? "Invalid code");
        return;
      }

      router.replace("/");
    } catch {
      setError("Something went wrong");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="flex min-h-svh items-center justify-center p-4">
      <Card className="w-full max-w-sm">
        {step === "password" ? (
          <>
            <CardHeader>
              <CardTitle>Reliva</CardTitle>
              <CardDescription>Enter your password to continue.</CardDescription>
            </CardHeader>
            <CardContent>
              <form onSubmit={handlePassword} className="flex flex-col gap-4">
                <div className="flex flex-col gap-1.5">
                  <Label htmlFor="password">Password</Label>
                  <Input
                    id="password"
                    type="password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    autoFocus
                    required
                  />
                </div>
                {error && (
                  <p className="text-sm text-destructive">{error}</p>
                )}
                <Button type="submit" disabled={loading}>
                  {loading ? "Checking…" : "Continue"}
                </Button>
              </form>
            </CardContent>
          </>
        ) : (
          <>
            <CardHeader>
              <CardTitle>{useBackup ? "Backup code" : "Two-factor auth"}</CardTitle>
              <CardDescription>
                {useBackup
                  ? "Enter one of your 12-character backup codes."
                  : "Enter the 6-digit code from your authenticator app."}
              </CardDescription>
            </CardHeader>
            <CardContent>
              <form onSubmit={handleTotp} className="flex flex-col gap-4">
                <div className="flex flex-col gap-1.5">
                  <Label htmlFor="code">
                    {useBackup ? "Backup code" : "Authenticator code"}
                  </Label>
                  <Input
                    id="code"
                    type="text"
                    inputMode={useBackup ? "text" : "numeric"}
                    value={code}
                    onChange={(e) => setCode(e.target.value)}
                    autoFocus
                    required
                  />
                </div>
                {error && (
                  <p className="text-sm text-destructive">{error}</p>
                )}
                <Button type="submit" disabled={loading}>
                  {loading ? "Verifying…" : "Sign in"}
                </Button>
                <button
                  type="button"
                  className="text-sm text-muted-foreground underline-offset-4 hover:underline"
                  onClick={() => {
                    setUseBackup((v) => !v);
                    setCode("");
                    setError("");
                  }}
                >
                  {useBackup
                    ? "Use authenticator app instead"
                    : "Use a backup code instead"}
                </button>
              </form>
            </CardContent>
          </>
        )}
      </Card>
    </div>
  );
}
