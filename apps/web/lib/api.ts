/**
 * Typed client for the Forum API.
 * Mirrors the backend envelope { success, data, error, meta? } and manages
 * access/refresh tokens with a single transparent refresh on 401.
 */

export const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

const ACCESS_KEY = "forum.accessToken";
const REFRESH_KEY = "forum.refreshToken";

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

/* ----------------------------- token storage ----------------------------- */

export const tokens = {
  access(): string | null {
    try {
      return localStorage.getItem(ACCESS_KEY);
    } catch {
      return null;
    }
  },
  refresh(): string | null {
    try {
      return localStorage.getItem(REFRESH_KEY);
    } catch {
      return null;
    }
  },
  set(access: string, refresh: string) {
    try {
      localStorage.setItem(ACCESS_KEY, access);
      localStorage.setItem(REFRESH_KEY, refresh);
    } catch {
      /* ignore storage failures */
    }
  },
  clear() {
    try {
      localStorage.removeItem(ACCESS_KEY);
      localStorage.removeItem(REFRESH_KEY);
    } catch {
      /* ignore */
    }
  },
};

/* ------------------------------- core fetch ------------------------------ */

interface RequestOptions {
  method?: string;
  body?: unknown;
  auth?: boolean;
  isForm?: boolean;
}

async function raw<T>(path: string, opts: RequestOptions, retry = true): Promise<{ data: T; meta?: ApiMeta }> {
  const url = path.startsWith("http") ? path : `${API_URL}${path}`;
  const headers: Record<string, string> = {};
  let body: BodyInit | undefined;

  if (opts.body !== undefined) {
    if (opts.isForm) {
      body = opts.body as FormData;
    } else {
      headers["Content-Type"] = "application/json";
      body = JSON.stringify(opts.body);
    }
  }
  if (opts.auth) {
    const token = tokens.access();
    if (token) headers["Authorization"] = `Bearer ${token}`;
  }

  let res: Response;
  try {
    res = await fetch(url, { method: opts.method ?? "GET", headers, body });
  } catch {
    throw new ApiRequestError(0, "network_error", "não foi possível conectar à API");
  }

  if (res.status === 401 && opts.auth && retry && tokens.refresh()) {
    const refreshed = await tryRefresh();
    if (refreshed) return raw<T>(path, opts, false);
  }

  let parsed: Envelope<T>;
  try {
    parsed = (await res.json()) as Envelope<T>;
  } catch {
    throw new ApiRequestError(res.status, "invalid_response", "resposta inválida da API");
  }

  if (!res.ok || !parsed.success) {
    throw new ApiRequestError(
      res.status,
      parsed.error?.code ?? "request_failed",
      parsed.error?.message ?? "a requisição falhou",
    );
  }
  return { data: parsed.data as T, meta: parsed.meta };
}

async function tryRefresh(): Promise<boolean> {
  const refreshToken = tokens.refresh();
  if (!refreshToken) return false;
  try {
    const { data } = await raw<TokenPair>("/api/v1/auth/refresh", {
      method: "POST",
      body: { refreshToken },
    }, false);
    tokens.set(data.accessToken, data.refreshToken);
    return true;
  } catch {
    tokens.clear();
    return false;
  }
}

/* --------------------------------- types --------------------------------- */

export interface PublicProfile {
  id: string;
  username: string;
  displayName: string;
  bio: string | null;
  avatarUrl: string | null;
  bannerUrl: string | null;
  reputation: number;
  status: string;
  createdAt: string;
  lastActiveAt: string | null;
}

export interface Account extends PublicProfile {
  email: string;
}

export interface TokenPair {
  accessToken: string;
  refreshToken: string;
  expiresIn: number;
}

export interface AuthResult {
  user: Account;
  roles: string[];
  permissions: string[];
  tokens: TokenPair;
}

export interface MeResult {
  user: Account;
  roles: string[];
  permissions: string[];
}

/* ------------------------------- API surface ----------------------------- */

export const api = {
  getHealth: () => raw<{ status: string }>("/health", {}),

  register: (input: { username: string; email: string; password: string; displayName?: string }) =>
    raw<AuthResult>("/api/v1/auth/register", { method: "POST", body: input }),

  login: (input: { identifier: string; password: string }) =>
    raw<AuthResult>("/api/v1/auth/login", { method: "POST", body: input }),

  me: () => raw<MeResult>("/api/v1/me", { auth: true }),

  logout: (refreshToken: string) =>
    raw<unknown>("/api/v1/auth/logout", { method: "POST", body: { refreshToken } }),

  publicProfile: (username: string) =>
    raw<PublicProfile>(`/api/v1/users/${encodeURIComponent(username)}`, {}),

  updateProfile: (input: { displayName: string; bio?: string | null }) =>
    raw<Account>("/api/v1/me", { method: "PATCH", body: input, auth: true }),

  changePassword: (input: { currentPassword: string; newPassword: string }) =>
    raw<unknown>("/api/v1/me/password", { method: "PUT", body: input, auth: true }),

  changeEmail: (input: { email: string; password: string }) =>
    raw<Account>("/api/v1/me/email", { method: "PUT", body: input, auth: true }),

  uploadAvatar: (file: File) => {
    const form = new FormData();
    form.append("file", file);
    return raw<Account>("/api/v1/me/avatar", { method: "POST", body: form, auth: true, isForm: true });
  },
};
