import { useState } from "react";
import { Navigate } from "react-router-dom";
import { usePostShortenURL } from "@/api/generated";
import { AccountMenu } from "@/components/account-menu";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useSession } from "@/lib/session";

const baseURL = window.location.origin;

export function Short() {
  const [url, setUrl] = useState("");
  const [result, setResult] = useState<string | null>(null);
  const { loading: checkingSession, authenticated: isAuthenticated } = useSession();

  const shortener = usePostShortenURL({
    request: { credentials: "include" },
  });

  if (checkingSession) {
    return (
      <div className="flex items-center justify-center min-h-screen p-4">
        <Card className="w-full max-w-md">
          <CardHeader>
            <CardTitle>Checking session...</CardTitle>
            <CardDescription>Please wait while we verify your login.</CardDescription>
          </CardHeader>
        </Card>
      </div>
    );
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  const { mutate, isPending, isError, error } = shortener;

  const handleSubmit = (e: React.SubmitEvent<HTMLFormElement>) => {
    e.preventDefault();

    mutate(
      {
        data: {
          original_url: url,
        },
      },
      {
        onSuccess: (response) => {
          if (response.status === 201) {
            setResult(`${baseURL}/${response.data.short_code}`);
          }
        },
        onError: (err) => {
          console.error("Error shortening URL:", err);
          setResult(null);
        }
      }
    );
  };

  return (
    <div className="flex items-center justify-center min-h-screen p-4">
      <AccountMenu />
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle>Shorten URL</CardTitle>
          <CardDescription>
            Enter a URL to get a shortened version
          </CardDescription>
        </CardHeader>

        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="url">URL</Label>
              <Input
                id="url"
                type="url"
                placeholder="https://example.com"
                value={url}
                onChange={(e) => setUrl(e.target.value)}
                required
              />
            </div>

            <Button type="submit" className="w-full" disabled={isPending}>
              {isPending ? "Shortening..." : "Shorten"}
            </Button>
          </form>

          {result && (
            <div className="mt-4 p-3 bg-muted rounded-md">
              <p className="text-sm font-medium">Shortened URL:</p>
              <a
                href={result}
                target="_blank"
                rel="noopener noreferrer"
                className="text-sm text-primary hover:underline break-all"
              >
                {result}
              </a>
            </div>
          )}

          {isError && (
            <p className="mt-4 text-sm text-destructive">
              Error: {(error as Error).message}
            </p>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

export default Short;
