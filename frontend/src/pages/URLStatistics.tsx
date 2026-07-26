import { useState } from "react";
import { Link, useParams } from "react-router-dom";
import {
  ArrowLeftIcon,
  ArrowSquareOutIcon,
  CalendarBlankIcon,
  ClockIcon,
  CursorClickIcon,
  LinkSimpleIcon,
  PencilSimpleIcon,
} from "@phosphor-icons/react";
import { useGetURLStatistics, useUpdateURLByID } from "@/api/generated";
import type { DtosDailyClickResponse } from "@/api/model";
import { APIError } from "@/api/fetcher";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

const baseURL = window.location.origin;

function formatTimestamp(value?: string) {
  return value ? new Date(value).toLocaleString() : "No clicks yet";
}

function formatChartDate(value: string, includeMonth = false) {
  return new Date(`${value}T00:00:00Z`).toLocaleDateString(undefined, {
    day: "numeric",
    month: includeMonth ? "short" : undefined,
    timeZone: "UTC",
  });
}

function LoadingCard() {
  return (
    <Card className="mx-auto w-full max-w-md">
      <CardHeader>
        <CardTitle>Loading analytics...</CardTitle>
        <CardDescription>We are preparing this link’s statistics.</CardDescription>
      </CardHeader>
    </Card>
  );
}

function StateCard({
  title,
  description,
}: {
  title: string;
  description: string;
}) {
  return (
    <Card className="mx-auto w-full max-w-lg text-center">
      <CardHeader>
        <CardTitle>{title}</CardTitle>
        <CardDescription>{description}</CardDescription>
      </CardHeader>
      <CardContent>
        <Button asChild variant="outline">
          <Link to="/myurls">
            <ArrowLeftIcon aria-hidden="true" />
            Back to my URLs
          </Link>
        </Button>
      </CardContent>
    </Card>
  );
}

function ClickChart({ days }: { days: DtosDailyClickResponse[] }) {
  const maximum = Math.max(1, ...days.map((day) => day.clicks ?? 0));

  return (
    <div className="overflow-x-auto pb-2">
      <div className="grid h-56 min-w-180 grid-cols-30 items-end gap-1.5" role="img" aria-label="Clicks during the last 30 days">
        {days.map((day, index) => {
          const clicks = day.clicks ?? 0;
          const height = clicks === 0 ? 2 : Math.max(8, (clicks / maximum) * 100);
          const showLabel = index === 0 || index === days.length - 1 || index % 7 === 0;

          return (
            <div key={day.date} className="flex h-full min-w-0 flex-col justify-end gap-2">
              <div className="flex flex-1 items-end">
                <div
                  className="w-full bg-primary/80 transition-colors hover:bg-primary"
                  style={{ height: `${height}%` }}
                  title={`${formatChartDate(day.date, true)}: ${clicks} ${clicks === 1 ? "click" : "clicks"}`}
                />
              </div>
              <span className="h-4 whitespace-nowrap text-center text-[10px] text-muted-foreground">
                {showLabel ? formatChartDate(day.date, true) : ""}
              </span>
            </div>
          );
        })}
      </div>
    </div>
  );
}

export function URLStatistics() {
  const { id = "" } = useParams<{ id: string }>();
  const [editingDestination, setEditingDestination] = useState(false);
  const [destination, setDestination] = useState("");
  const { data, isLoading, error, refetch } = useGetURLStatistics(id, {
    request: { credentials: "include" },
    query: { enabled: Boolean(id) },
  });
  const destinationUpdater = useUpdateURLByID({
    request: { credentials: "include" },
  });

  const isNotFound = error instanceof APIError && error.status === 404;

  if (isLoading) {
    return (
      <main className="min-h-screen bg-muted/30 px-4 py-16 sm:px-6">
        <LoadingCard />
      </main>
    );
  }

  if (error || !data) {
    return (
      <main className="min-h-screen bg-muted/30 px-4 py-16 sm:px-6">
        <StateCard
          title={isNotFound ? "Link not found" : "Could not load analytics"}
          description={
            isNotFound
              ? "This URL does not exist or is not available to your account."
              : "Please try loading the analytics again in a moment."
          }
        />
      </main>
    );
  }

  const shortURL = `${baseURL}/${data.url.short_code}`;
  const totalClicks = data.total_clicks ?? 0;
  const topReferrerMaximum = Math.max(1, ...data.top_referrers.map((item) => item.clicks ?? 0));

  const startEditingDestination = () => {
    setDestination(data.url.original_url);
    destinationUpdater.reset();
    setEditingDestination(true);
  };

  const handleDestinationSubmit = (event: React.SubmitEvent<HTMLFormElement>) => {
    event.preventDefault();
    destinationUpdater.mutate(
      {
        id,
        data: { original_url: destination },
      },
      {
        onSuccess: () => {
          setEditingDestination(false);
          void refetch();
        },
      },
    );
  };

  return (
    <main className="min-h-screen bg-muted/30 px-4 py-16 sm:px-6">
      <div className="mx-auto w-full max-w-6xl space-y-6">
        <div>
          <Button asChild variant="ghost" className="-ml-2 mb-4">
            <Link to="/myurls">
              <ArrowLeftIcon aria-hidden="true" />
              Back to my URLs
            </Link>
          </Button>
          <div className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
            <div className="min-w-0 space-y-1">
              <p className="text-sm font-medium text-primary">Link analytics</p>
              <h1 className="text-3xl font-semibold tracking-tight">Performance overview</h1>
              <a
                href={data.url.original_url}
                target="_blank"
                rel="noreferrer"
                className="block max-w-2xl truncate text-muted-foreground hover:text-foreground hover:underline"
                title={data.url.original_url}
              >
                {data.url.original_url}
              </a>
            </div>
            <Button asChild variant="outline">
              <a href={shortURL} target="_blank" rel="noreferrer">
                <ArrowSquareOutIcon aria-hidden="true" />
                Open short link
              </a>
            </Button>
          </div>
        </div>

        <Card>
          <CardContent className="flex flex-col gap-3 py-4 sm:flex-row sm:items-center sm:justify-between">
            <div className="flex min-w-0 items-center gap-2">
              <LinkSimpleIcon className="shrink-0 text-primary" aria-hidden="true" />
              <span className="truncate font-mono text-sm">{shortURL}</span>
            </div>
            <Badge variant="secondary">Created {formatTimestamp(data.url.created_at)}</Badge>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex-row items-start justify-between gap-4">
            <div className="min-w-0">
              <CardTitle>Destination URL</CardTitle>
              <CardDescription>Change where this short link redirects without changing its short code.</CardDescription>
            </div>
            {!editingDestination && (
              <Button variant="outline" size="sm" onClick={startEditingDestination}>
                <PencilSimpleIcon aria-hidden="true" />
                Edit
              </Button>
            )}
          </CardHeader>
          <CardContent>
            {editingDestination ? (
              <form onSubmit={handleDestinationSubmit} className="space-y-3">
                <div className="space-y-1">
                  <Input
                    id="destination-url"
                    type="url"
                    value={destination}
                    onChange={(event) => setDestination(event.target.value)}
                    required
                    autoFocus
                  />
                </div>
                {destinationUpdater.isError && (
                  <p className="text-sm text-destructive">
                    {(destinationUpdater.error as unknown as Error).message}
                  </p>
                )}
                <div className="flex gap-2">
                  <Button type="submit" disabled={destinationUpdater.isPending}>
                    {destinationUpdater.isPending ? "Saving..." : "Save destination"}
                  </Button>
                  <Button
                    type="button"
                    variant="outline"
                    disabled={destinationUpdater.isPending}
                    onClick={() => setEditingDestination(false)}
                  >
                    Cancel
                  </Button>
                </div>
              </form>
            ) : (
              <a
                href={data.url.original_url}
                target="_blank"
                rel="noreferrer"
                className="block truncate text-sm text-primary hover:underline"
                title={data.url.original_url}
              >
                {data.url.original_url}
              </a>
            )}
          </CardContent>
        </Card>

        <div className="grid gap-4 md:grid-cols-3">
          <Card>
            <CardHeader className="pb-2">
              <CardDescription className="flex items-center gap-2">
                <CursorClickIcon aria-hidden="true" />
                Total clicks
              </CardDescription>
              <CardTitle className="text-3xl">{totalClicks.toLocaleString()}</CardTitle>
            </CardHeader>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardDescription className="flex items-center gap-2">
                <ClockIcon aria-hidden="true" />
                Last click
              </CardDescription>
              <CardTitle className="text-lg">{formatTimestamp(data.last_clicked_at)}</CardTitle>
            </CardHeader>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardDescription className="flex items-center gap-2">
                <CalendarBlankIcon aria-hidden="true" />
                Tracking window
              </CardDescription>
              <CardTitle className="text-lg">Last 30 days</CardTitle>
            </CardHeader>
          </Card>
        </div>

        {totalClicks === 0 && (
          <Card className="border-dashed">
            <CardHeader>
              <CardTitle>No clicks yet</CardTitle>
              <CardDescription>Share the short link and its first visit will appear here.</CardDescription>
            </CardHeader>
          </Card>
        )}

        <Card>
          <CardHeader>
            <CardTitle>Clicks over time</CardTitle>
            <CardDescription>Daily visits during the latest 30 UTC calendar days.</CardDescription>
          </CardHeader>
          <CardContent>
            <ClickChart days={data.clicks_by_day} />
          </CardContent>
        </Card>

        <div className="grid gap-6 lg:grid-cols-[2fr_3fr]">
          <Card>
            <CardHeader>
              <CardTitle>Top referrers</CardTitle>
              <CardDescription>Where visitors opened this link.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              {data.top_referrers.length === 0 ? (
                <p className="text-sm text-muted-foreground">No referrer data yet.</p>
              ) : (
                data.top_referrers.map((item) => (
                  <div key={item.referrer} className="space-y-2">
                    <div className="flex items-center justify-between gap-3 text-sm">
                      <span className="truncate font-medium">{item.referrer}</span>
                      <span className="text-muted-foreground">{item.clicks ?? 0}</span>
                    </div>
                    <div className="h-1.5 bg-muted">
                      <div
                        className="h-full bg-primary"
                        style={{ width: `${((item.clicks ?? 0) / topReferrerMaximum) * 100}%` }}
                      />
                    </div>
                  </div>
                ))
              )}
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Recent activity</CardTitle>
              <CardDescription>The ten latest visits to this link.</CardDescription>
            </CardHeader>
            <CardContent className="p-0">
              <div className="overflow-x-auto">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Clicked at</TableHead>
                      <TableHead>Referrer</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {data.recent_clicks.length === 0 ? (
                      <TableRow>
                        <TableCell colSpan={2} className="h-24 text-center text-muted-foreground">
                          No recent clicks.
                        </TableCell>
                      </TableRow>
                    ) : (
                      data.recent_clicks.map((click, index) => (
                        <TableRow key={`${click.clicked_at}-${index}`}>
                          <TableCell className="whitespace-nowrap">{formatTimestamp(click.clicked_at)}</TableCell>
                          <TableCell>{click.referrer}</TableCell>
                        </TableRow>
                      ))
                    )}
                  </TableBody>
                </Table>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </main>
  );
}

export default URLStatistics;
