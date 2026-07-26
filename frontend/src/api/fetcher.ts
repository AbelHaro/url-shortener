type APIRequest = {
  url: string;
  method: string;
  headers?: HeadersInit;
  data?: unknown;
  signal?: AbortSignal;
};

export class APIError extends Error {
  status: number;

  constructor(message: string, status: number) {
    super(message);
    this.name = "APIError";
    this.status = status;
  }
}

export async function authFetch<T>(request: APIRequest, init?: RequestInit): Promise<T> {
  const response = await fetch(
    `${import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8080"}${request.url}`,
    {
      ...init,
      method: request.method,
      headers: request.headers,
      body: request.data === undefined ? undefined : JSON.stringify(request.data),
      signal: request.signal,
      credentials: "include",
    },
  );

  const body = await response.text();
  const data = body ? JSON.parse(body) : undefined;

  if (!response.ok) {
    throw new APIError(data?.error ?? `Request failed with status ${response.status}`, response.status);
  }

  return data as T;
}
