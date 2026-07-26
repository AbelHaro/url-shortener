import { Navigate, Outlet } from "react-router-dom";
import { AccountMenu } from "@/components/account-menu";
import { Card, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { useSession } from "@/lib/session";

export function AuthenticatedLayout() {
  const { loading, authenticated, user } = useSession();

  if (loading) {
    return (
      <main className="flex min-h-screen items-center justify-center p-4">
        <Card className="w-full max-w-md">
          <CardHeader>
            <CardTitle>Checking session...</CardTitle>
            <CardDescription>Please wait while we verify your login.</CardDescription>
          </CardHeader>
        </Card>
      </main>
    );
  }

  if (!authenticated) {
    return <Navigate to="/login" replace />;
  }

  return (
    <>
      <AccountMenu user={user} />
      <Outlet />
    </>
  );
}
