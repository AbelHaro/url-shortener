import { useState } from "react";
import { useShortenUrl } from "@/api/hooks/useUrl";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

const baseURL = window.location.origin;

export function Short() {
  const [url, setUrl] = useState("");
  const [result, setResult] = useState<string | null>(null);
  const shortenUrl = useShortenUrl();

  const handleSubmit = (e: React.SubmitEvent) => {
    e.preventDefault();
    shortenUrl.mutate(url, {
      onSuccess: (data) => {
        const fullShortUrl = `${baseURL}/${data.shortCode}`;
        setResult(fullShortUrl);
      },
    });
  };

  return (
    <div className="flex items-center justify-center min-h-screen p-4">
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
            <Button type="submit" className="w-full" disabled={shortenUrl.isPending}>
              {shortenUrl.isPending ? "Shortening..." : "Shorten"}
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
          {shortenUrl.isError && (
            <p className="mt-4 text-sm text-destructive">
              Error: {(shortenUrl.error as Error).message}
            </p>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

export default Short;
