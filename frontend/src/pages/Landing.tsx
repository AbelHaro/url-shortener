import { Link } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";

export function Landing() {
  return (
    <div className="flex items-center justify-center min-h-screen p-4">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle>URL Shortener</CardTitle>
          <CardDescription>
            A overengineered and fast URL shortening service for learning purposes.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <p className="text-sm text-muted-foreground">
            This application allows you to create short, memorable URLs from long web addresses.
            Simply enter your URL and get a compact version that's easy to share.
          </p>
          <div className="space-y-2">
            <p className="text-sm font-medium">Features:</p>
            <ul className="text-sm text-muted-foreground list-disc list-inside">
              <li>Instant URL shortening</li>
              <li>Easy to remember short codes</li>
              <li>Fast redirects</li>
            </ul>
          </div>
          <Button asChild className="w-full">
            <Link to="/login">Login to shorten a URL</Link>
          </Button>
          <Button asChild variant="outline" className="w-full">
            <Link to="/register">Create an account</Link>
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}

export default Landing;
