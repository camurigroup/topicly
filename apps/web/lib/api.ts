/**
 * Typed client for the Forum API.
 * Mirrors the backend response envelope: { success, data, error, meta? }.
 */

export const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export interface ApiMeta {
  total: number;
  page: number;
  limit: number;
}

export interface ApiError {
  code: string;
  message: string;
}

export interface Envelope<T> {
  success: boolean;
  data: T | null;
  error: ApiError | null;
  meta?: ApiMeta;
}

export class ApiRequestError extends Error {
  constructor(
    public readonly status: number,
    public readonly code: string,
    message: string,
  ) {
    super(message);
    this.name = "ApiRequestError";
  }
}

/**
 * Perform a request against the API and unwrap the envelope.
 * Throws ApiRequestError on a non-success envelope or transport failure.
 */
export async function apiFetch<T>(
  path: string,
  init?: RequestInit,
): Promise<{ data: T; meta?: ApiMeta }> {
  const url = path.startsWith("http") ? path : `${API_URL}${path}`;

  let res: Response;
  try {
    res = await fetch(url, {
      ...init,
      headers: {
        "Content-Type": "application/json",
        ...(init?.headers ?? {}),
      },
    });
  } catch (cause) {
    throw new ApiRequestError(0, "network_error", "could not reach the API");
  }

  let body: Envelope<T>;
  try {
    body = (await res.json()) as Envelope<T>;
  } catch {
    throw new ApiRequestError(res.status, "invalid_response", "invalid API response");
  }

  if (!res.ok || !body.success || body.data === null) {
    throw new ApiRequestError(
      res.status,
      body.error?.code ?? "request_failed",
      body.error?.message ?? "request failed",
    );
  }

  return { data: body.data, meta: body.meta };
}

/** Health status returned by GET /health. */
export interface HealthStatus {
  status: string;
}

export function getHealth() {
  return apiFetch<HealthStatus>("/health");
}
