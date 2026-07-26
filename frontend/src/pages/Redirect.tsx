import { useEffect, useRef, useState } from "react";
import { useParams } from "react-router-dom";
import { resolveShortURL } from "@/api/generated";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";

export function Redirect() {
  const { shortCode } = useParams<{ shortCode: string }>();
  const attempted = useRef(false);
  const [isError, setIsError] = useState(false);

  useEffect(() => {
    if (!shortCode || attempted.current) {
      return;
    }
    attempted.current = true;

    void resolveShortURL(shortCode, { referrer: document.referrer })
      .then((response) => {
        window.location.replace(response.original_url);
      })
      .catch(() => {
        setIsError(true);
      });
  }, [shortCode]);

  if (isError || !shortCode) {
    return (
      <div className="flex min-h-screen items-center justify-center p-4">
        <Card className="w-full max-w-md">
          <CardHeader>
            <CardTitle>URL Not Found</CardTitle>
            <CardDescription>The short URL you are trying to access does not exist.</CardDescription>
          </CardHeader>
        </Card>
      </div>
    );
  }

  return (
    <div className="flex min-h-screen items-center justify-center p-4">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle>Redirecting...</CardTitle>
          <CardDescription>Resolving your destination.</CardDescription>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">You will be redirected automatically.</p>
        </CardContent>
      </Card>
    </div>
  );
}

export default Redirect;
