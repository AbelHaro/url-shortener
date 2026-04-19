import { useEffect } from "react";
import { useParams } from "react-router-dom";
import { useGetURLByShortCode } from "@/api/generated";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";

export function Redirect() {
  const { shortCode } = useParams<{ shortCode: string }>();
  const { data, isLoading, isError } = useGetURLByShortCode(shortCode!, {
    query: {
    enabled: !!shortCode,
  }
});


  useEffect(() => {
    if (data?.status === 200 && data.data.original_url) {
      window.location.href = data.data.original_url;
    }
  }, [data]);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen p-4">
        <Card className="w-full max-w-md">
          <CardHeader>
            <CardTitle>Redirecting...</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">Please wait while we redirect you.</p>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (isError) {
    return (
      <div className="flex items-center justify-center min-h-screen p-4">
        <Card className="w-full max-w-md">
          <CardHeader>
            <CardTitle>URL Not Found</CardTitle>
            <CardDescription>
              The short URL you are trying to access does not exist.
            </CardDescription>
          </CardHeader>
        </Card>
      </div>
    );
  }

  return null;
}

export default Redirect;
