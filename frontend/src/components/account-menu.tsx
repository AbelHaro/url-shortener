import { SignOutIcon } from "@phosphor-icons/react";

import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Button } from "@/components/ui/button";
import { useSession } from "@/lib/session";

const apiBaseURL = import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8080";

export function AccountMenu() {
  const { authenticated, user } = useSession();

  const handleLogout = async () => {
    await fetch(`${apiBaseURL}/auth/logout`, {
      method: "POST",
      credentials: "include",
    });

    window.location.href = "/";
  };

  if (!authenticated) return null;

  return (
    <div className="fixed right-4 top-4 z-50">
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="outline" className="max-w-40 truncate">
            {user?.name ?? "Account"}
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuItem onClick={handleLogout} className="gap-2">
            <SignOutIcon />
            Logout
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  );
}
