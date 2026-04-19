import { useState } from "react";
import { Link, Navigate, useNavigate } from "react-router-dom";

import { usePostAuthAnonymousRegister, usePostAuthRegister } from "@/api/generated";
import { AccountMenu } from "@/components/account-menu";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useSession } from "@/lib/session";
import { toast } from "sonner";

export function Register() {
  const navigate = useNavigate();
  const [mode, setMode] = useState<"email" | "anonymous">("email");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [message, setMessage] = useState<string | null>(null);

  const registerMutation = usePostAuthRegister();
  const anonymousMutation = usePostAuthAnonymousRegister();
  const { loading, authenticated } = useSession();

  const busy = registerMutation.isPending || anonymousMutation.isPending;
  const error = registerMutation.error ?? anonymousMutation.error;
  const isError = registerMutation.isError || anonymousMutation.isError;

  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center p-4">
        <Card className="w-full max-w-md">
          <CardHeader>
            <CardTitle>Checking session...</CardTitle>
            <CardDescription>Please wait while we verify your login.</CardDescription>
          </CardHeader>
        </Card>
      </div>
    );
  }

  if (authenticated) {
    return <Navigate to="/short" replace />;
  }

  const handleEmailRegister = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setMessage(null);

    if (password !== confirmPassword) {
      setMessage("Passwords do not match.");
      return;
    }

    registerMutation.mutate(
      {
        data: {
          email,
          password,
        },
      },
      {
        onSuccess: (response) => {
          if (response.status === 409) {
            const errorMessage = "Email already in use";
            toast.error(errorMessage);
            setMessage(errorMessage);
            return;
          }

          if (response.status !== 201) return;

          const { user } = response.data;
          if (!user?.name) return;

          toast.success(`Account created for ${user.name}`);
          setMessage(`Welcome, ${user.name}`);
          navigate("/short");
        },
      }
    );
  };

  const handleAnonymousRegister = () => {
    setMessage(null);

    anonymousMutation.mutate(undefined, {
        onSuccess: (response) => {
          if (response.status !== 201) return;

          const { user } = response.data;
          if (!user?.name) return;

          toast.success(`Anonymous account created for ${user.name}`);
          setMessage(`Anonymous account created: ${user.name}`);
          navigate("/short");
        },
    });
  };

  return (
    <div className="flex min-h-screen items-center justify-center p-4">
      <AccountMenu />
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle>Create account</CardTitle>
          <CardDescription>
            Register with email or create an anonymous account.
          </CardDescription>
        </CardHeader>

        <CardContent className="space-y-4">
          <div className="flex gap-2">
            <Button
              type="button"
              variant={mode === "email" ? "default" : "outline"}
              className="flex-1"
              onClick={() => setMode("email")}
            >
              Email
            </Button>
            <Button
              type="button"
              variant={mode === "anonymous" ? "default" : "outline"}
              className="flex-1"
              onClick={() => setMode("anonymous")}
            >
              Anonymous
            </Button>
          </div>

          {mode === "email" ? (
            <form onSubmit={handleEmailRegister} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="email">Email</Label>
                <Input
                  id="email"
                  type="email"
                  placeholder="you@example.com"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  autoComplete="email"
                  required
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="password">Password</Label>
                <Input
                  id="password"
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  autoComplete="new-password"
                  minLength={8}
                  required
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="confirmPassword">Confirm password</Label>
                <Input
                  id="confirmPassword"
                  type="password"
                  value={confirmPassword}
                  onChange={(e) => setConfirmPassword(e.target.value)}
                  autoComplete="new-password"
                  minLength={8}
                  required
                />
              </div>

              <Button type="submit" className="w-full" disabled={busy}>
                {busy ? "Creating account..." : "Register"}
              </Button>
            </form>
          ) : (
            <div className="space-y-4">
              <p className="text-sm text-muted-foreground">
                Create an anonymous account. You will get a name and tokens right away.
              </p>
              <Button type="button" className="w-full" onClick={handleAnonymousRegister} disabled={busy}>
                {busy ? "Creating anonymous account..." : "Continue anonymously"}
              </Button>
            </div>
          )}

          {message && <p className="text-sm text-red-600">{message}</p>}
          {isError && <p className="text-sm text-destructive">Error: {(error as Error).message}</p>}
        </CardContent>

        <CardFooter className="justify-center">
          <Button variant="link" asChild>
            <Link to="/login">Already have an account? Login</Link>
          </Button>
        </CardFooter>
      </Card>
    </div>
  );
}

export default Register;
