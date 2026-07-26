import { useState } from "react";
import { Link } from "react-router-dom";
import { ChartLineUpIcon, DotsThreeVerticalIcon, LinkSimpleIcon, PlusIcon, TrashIcon } from "@phosphor-icons/react";
import type { ColumnDef } from "@tanstack/react-table";
import { useDeleteURLByID, useGetURLsByUserID } from "@/api/generated";
import type { DomainURL } from "@/api/model";
import { DataTable } from "@/components/data-table";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

const baseURL = window.location.origin;

type URLRow = Omit<DomainURL, "created_at" | "updated_at"> & {
  created_at: string;
  updated_at: string;
};

function timestamp(value?: string) {
  return value ? new Date(value).getTime() : 0;
}

function formatDate(value?: string) {
  return value ? new Date(value).toLocaleString() : "—";
}

function LinkActions({
  url,
  onDelete,
  isDeleting,
}: {
  url: URLRow;
  onDelete: (id: string) => void;
  isDeleting: boolean;
}) {
  const [confirming, setConfirming] = useState(false);
  const canDelete = Boolean(url.id);

  return (
    <AlertDialog open={confirming} onOpenChange={setConfirming}>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="ghost" size="icon-sm" aria-label="Link actions">
            <DotsThreeVerticalIcon aria-hidden="true" size={18} weight="bold" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuItem asChild disabled={!url.id}>
            <Link to={url.id ? `/myurls/${url.id}` : "/myurls"}>
              <ChartLineUpIcon aria-hidden="true" />
              Analytics
            </Link>
          </DropdownMenuItem>
          <DropdownMenuItem
            variant="destructive"
            disabled={!canDelete || isDeleting}
            onSelect={() => setConfirming(true)}
          >
            <TrashIcon aria-hidden="true" />
            Delete
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Delete this link?</AlertDialogTitle>
          <AlertDialogDescription>
            This will permanently delete the shortened link for {url.original_url ?? "this destination"}.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={isDeleting}>Cancel</AlertDialogCancel>
          <AlertDialogAction
            variant="destructive"
            disabled={!url.id || isDeleting}
            onClick={() => url.id && onDelete(url.id)}
          >
            {isDeleting ? "Deleting..." : "Delete link"}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}

function getColumns(onDelete: (id: string) => void, isDeleting: boolean): ColumnDef<URLRow>[] {
  return [
  {
    accessorKey: "original_url",
    header: "Destination",
    cell: ({ row }) => {
      const originalURL = row.original.original_url;

      return originalURL ? (
        <a
          href={originalURL}
          target="_blank"
          rel="noreferrer"
          className="block max-w-md truncate font-medium text-foreground hover:text-primary hover:underline"
          title={originalURL}
        >
          {originalURL}
        </a>
      ) : (
        "—"
      );
    },
  },
  {
    accessorKey: "short_code",
    header: "Short link",
    cell: ({ row }) => {
      const shortCode = row.original.short_code;
      const shortURL = shortCode ? `${baseURL}/${shortCode}` : undefined;

      return shortURL ? (
        <a
          href={shortURL}
          target="_blank"
          rel="noreferrer"
          className="inline-flex items-center gap-1.5 font-mono text-sm text-primary hover:underline"
        >
          <LinkSimpleIcon aria-hidden="true" size={16} />
          {shortCode}
        </a>
      ) : (
        "—"
      );
    },
  },
  {
    accessorKey: "created_at",
    header: "Created",
  },
  {
    accessorKey: "updated_at",
    header: "Last updated",
  },
  {
    id: "actions",
    header: () => <span className="sr-only">Actions</span>,
    cell: ({ row }) => <LinkActions url={row.original} onDelete={onDelete} isDeleting={isDeleting} />,
  },
];
}

function PageState({ title, description }: { title: string; description: string }) {
  return (
    <Card className="w-full max-w-lg text-center">
      <CardHeader className="items-center">
        <CardTitle>{title}</CardTitle>
        <CardDescription>{description}</CardDescription>
      </CardHeader>
      <CardContent>
        <Button asChild>
          <Link to="/short">
            <PlusIcon aria-hidden="true" />
            Shorten a URL
          </Link>
        </Button>
      </CardContent>
    </Card>
  );
}

export function MyURLs() {
  const { data, isLoading, isError, refetch } = useGetURLsByUserID({
    request: { credentials: "include" },
  });
  const { mutate: deleteURL, isPending: isDeleting } = useDeleteURLByID({
    request: { credentials: "include" },
  });

  const handleDelete = (id: string) => {
    deleteURL(
      { id },
      {
        onSuccess: () => {
          void refetch();
        },
      }
    );
  };

  const urls = data ?? [];
  const formattedURLs = [...urls]
    .sort((a, b) => timestamp(b.created_at) - timestamp(a.created_at))
    .map((url) => ({
      ...url,
      created_at: formatDate(url.created_at),
      updated_at: formatDate(url.updated_at),
    }));

  return (
    <main className="min-h-screen bg-muted/30 px-4 py-16 sm:px-6">
      <div className="mx-auto w-full max-w-6xl">
        <div className="mb-8 flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
          <div className="space-y-1">
            <p className="text-sm font-medium text-primary">Your workspace</p>
            <h1 className="text-3xl font-semibold tracking-tight">My URLs</h1>
            <p className="text-muted-foreground">Manage the links you have shortened.</p>
          </div>
          <Button asChild>
            <Link to="/short">
              <PlusIcon aria-hidden="true" />
              Shorten a URL
            </Link>
          </Button>
        </div>

        {isLoading ? (
          <PageState title="Loading your URLs..." description="We are fetching your shortened links." />
        ) : isError ? (
          <PageState title="Could not load your URLs" description="Please try again in a moment." />
        ) : urls.length === 0 ? (
          <PageState title="No shortened URLs yet" description="Create your first short link to see it here." />
        ) : (
          <Card>
            <CardHeader className="border-b">
              <div className="flex items-center justify-between gap-4">
                <div>
                  <CardTitle>Your links</CardTitle>
                  <CardDescription>Newest links appear first.</CardDescription>
                </div>
                <Badge variant="secondary">
                  {urls.length} {urls.length === 1 ? "link" : "links"}
                </Badge>
              </div>
            </CardHeader>
            <CardContent className="p-0">
              <div className="overflow-x-auto">
                <DataTable columns={getColumns(handleDelete, isDeleting)} data={formattedURLs} />
              </div>
            </CardContent>
          </Card>
        )}
      </div>
    </main>
  );
}

export default MyURLs;
