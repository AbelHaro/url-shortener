import { useState } from "react";
import { Link, Navigate, useNavigate } from "react-router-dom";

import { usePostAuthLogin } from "@/api/generated";
import { AccountMenu } from "@/components/account-menu";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useSession } from "@/lib/session";
import { toast } from "sonner";

export function Login() {
  const navigate = useNavigate();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [message, setMessage] = useState<string | null>(null);
  const { loading, authenticated } = useSession();

  const loginMutation = usePostAuthLogin();

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

  const handleLogin = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setMessage(null);

    loginMutation.mutate(
      {
        data: {
          email,
          password,
        },
      },
      {
        onSuccess: (response) => {
          if (response.status !== 200) return;

          const { user } = response.data;
          if (!user?.name) return;

          toast.success(`Welcome back, ${user.name}`);
          setMessage(`Welcome back, ${user.name}`);
          navigate("/short");
        },
      }
    );
  };

  return (
    <div className="flex min-h-screen items-center justify-center p-4">
      <AccountMenu />
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle>Login</CardTitle>
          <CardDescription>Enter your email and password to continue.</CardDescription>
        </CardHeader>

        <CardContent className="space-y-4">
          <form onSubmit={handleLogin} className="space-y-4">
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
                autoComplete="current-password"
                required
              />
            </div>

            <Button type="submit" className="w-full" disabled={loginMutation.isPending}>
              {loginMutation.isPending ? "Logging in..." : "Login"}
            </Button>
          </form>

          {message && <p className="text-sm text-green-600">{message}</p>}
          {loginMutation.isError && <p className="text-sm text-destructive">Error: {(loginMutation.error as Error).message}</p>}
        </CardContent>

        <CardFooter className="justify-center">
          <Button variant="link" asChild>
            <Link to="/register">Need an account? Register</Link>
          </Button>
        </CardFooter>
      </Card>
    </div>
  );
}

export default Login;
