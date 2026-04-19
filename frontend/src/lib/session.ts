import { useEffect, useState } from "react";

const apiBaseURL = import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8080";

export type SessionUser = {
  id: string;
  email: string;
  name: string;
};

export function useSession() {
  const [loading, setLoading] = useState(true);
  const [authenticated, setAuthenticated] = useState(false);
  const [user, setUser] = useState<SessionUser | null>(null);

  useEffect(() => {
    const checkSession = async () => {
      try {
        const response = await fetch(`${apiBaseURL}/auth/session`, {
          credentials: "include",
        });

        setAuthenticated(response.ok);
        if (response.ok) {
          const data = (await response.json()) as { user?: SessionUser };
          setUser(data.user ?? null);
        } else {
          setUser(null);
        }
      } finally {
        setLoading(false);
      }
    };

    void checkSession();
  }, []);

  return { loading, authenticated, user };
}
