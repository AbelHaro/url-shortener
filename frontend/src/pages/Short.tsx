import { useState } from "react";
import { usePostShortenURL } from "@/api/generated";
import { QRCode } from "@/components/qr-code";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { CopyIcon } from "@phosphor-icons/react";
import { toast } from "sonner";


const baseURL = window.location.origin;

export function Short() {
  const [url, setUrl] = useState("");
  const [result, setResult] = useState<string | null>(null);

  const shortener = usePostShortenURL({
    request: { credentials: "include" },
  });

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
          setResult(`${baseURL}/${response.short_code}`);
        },
        onError: (err) => {
          console.error("Error shortening URL:", err);
          setResult(null);
        }
      }
    );
  };

  const handleCopyToClipboard = async () => {
    if (result) {
      await navigator.clipboard.writeText(result);
      toast.success("Copied to clipboard!", {
        description: "The short link has been copied to your clipboard.",
        duration: 3000,
        action: {
          label: "Dismiss",
          onClick: () => toast.dismiss(),
        },
      });
    }
  };

  return (
    <main className="flex min-h-screen items-center justify-center bg-muted/30 p-4 sm:p-6">
      <Card className="w-full max-w-lg overflow-hidden shadow-sm">
        <CardHeader className="border-b bg-card pb-5">
          <CardTitle>Shorten a link</CardTitle>
          <CardDescription>
            Paste a long URL to create a compact, shareable link.
          </CardDescription>
        </CardHeader>

        <CardContent className="space-y-5 pt-6">
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="url">Destination URL</Label>
              <Input
                id="url"
                type="url"
                placeholder="https://example.com"
                value={url}
                onChange={(e) => setUrl(e.target.value)}
                required
                className="h-10"
              />
            </div>

            <Button type="submit" className="h-10 w-full" disabled={isPending}>
              {isPending ? "Shortening..." : "Shorten"}
            </Button>
          </form>

          {result && (
            <section
              className="flex flex-col items-center gap-4 rounded-lg border bg-muted/40 p-4"
              aria-live="polite"
            >
              <div className="w-full space-y-1 text-center">
                <p className="text-sm font-medium">Your short link is ready:</p>
                <div className="flex flex-row items-center justify-center gap-2">
                  <a
                    href={result}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="block break-all text-sm font-medium text-primary underline-offset-4 hover:underline"
                  >
                    {result}
                  </a>
                  <Button
                    type="button"
                    variant="secondary"
                    size="icon"
                    onClick={handleCopyToClipboard}
                    aria-label="Copy short link"
                  >
                    <CopyIcon size={16} aria-hidden="true" />
                  </Button>
                </div>
              </div>
              <QRCode value={result} />
            </section>
          )}

          {isError && (
            <p className="rounded-md border border-destructive/30 bg-destructive/10 p-3 text-sm text-destructive" role="alert">
              Error: {(error as Error).message}
            </p>
          )}
        </CardContent>
      </Card>
    </main>
  );
}

export default Short;
