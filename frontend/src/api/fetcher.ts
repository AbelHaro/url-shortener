export async function authFetch<T>(input: RequestInfo | URL, init?: RequestInit): Promise<T> {
  const response = await fetch(input, {
    ...init,
    credentials: "include",
  });

  const body = await response.text();
  const data = body ? JSON.parse(body) : undefined;

  return {
    data,
    status: response.status,
    headers: response.headers,
  } as T;
}
