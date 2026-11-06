// Thin fetch wrapper around the Go backend. Every request includes cookies so
// the session is sent cross-origin (frontend :3000 -> backend :8080).

export const API_URL =
  process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export class ApiError extends Error {
  status: number;
  constructor(message: string, status: number) {
    super(message);
    this.status = status;
  }
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const res = await fetch(API_URL + path, {
    credentials: "include",
    ...options,
  });
  if (!res.ok) {
    let message = res.statusText;
    try {
      const body = await res.json();
      if (body?.error) message = body.error;
    } catch {
      /* ignore non-JSON error bodies */
    }
    throw new ApiError(message, res.status);
  }
  if (res.status === 204) return undefined as T;
  const text = await res.text();
  return text ? (JSON.parse(text) as T) : (undefined as T);
}

export const api = {
  get: <T>(path: string) => request<T>(path),
  post: <T>(path: string, body?: unknown) =>
    request<T>(path, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: body === undefined ? undefined : JSON.stringify(body),
    }),
  put: <T>(path: string, body?: unknown) =>
    request<T>(path, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: body === undefined ? undefined : JSON.stringify(body),
    }),
  // form posts multipart/form-data (used for uploads).
  form: <T>(path: string, data: FormData) =>
    request<T>(path, { method: "POST", body: data }),
};

// mediaUrl turns a stored image path (e.g. "/uploads/x.png") into an absolute
// URL served by the backend.
export function mediaUrl(path: string | null | undefined): string | undefined {
  if (!path) return undefined;
  return API_URL + path;
}

// wsUrl is the websocket endpoint derived from the API URL.
export function wsUrl(): string {
  return API_URL.replace(/^http/, "ws") + "/ws";
}
