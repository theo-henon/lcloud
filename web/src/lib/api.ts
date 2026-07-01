export type UserRole = "admin" | "user";

export interface User {
  id: string;
  email: string;
  role: UserRole;
  created_at: string;
}

export interface LoginResponse {
  access_token: string;
  refresh_token: string;
  user: User;
}

export interface RefreshResponse {
  access_token: string;
  refresh_token: string;
}

export class ApiError extends Error {
  constructor(
    message: string,
    readonly status: number,
    readonly code?: string,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

type RequestOptions = RequestInit & {
  skipAuth?: boolean;
};

let accessTokenProvider: (() => string | null) | null = null;
let refreshHandler: (() => Promise<boolean>) | null = null;

export function configureApiAuth(
  getAccessToken: () => string | null,
  refreshTokens: () => Promise<boolean>,
) {
  accessTokenProvider = getAccessToken;
  refreshHandler = refreshTokens;
}

async function parseError(response: Response): Promise<ApiError> {
  try {
    const body = (await response.json()) as { error?: string; code?: string };
    return new ApiError(body.error ?? "Request failed", response.status, body.code);
  } catch {
    return new ApiError("Request failed", response.status);
  }
}

export async function apiRequest<T>(
  path: string,
  options: RequestOptions = {},
): Promise<T> {
  const headers = new Headers(options.headers);
  if (!headers.has("Content-Type") && options.body) {
    headers.set("Content-Type", "application/json");
  }

  if (!options.skipAuth && accessTokenProvider) {
    const token = accessTokenProvider();
    if (token) {
      headers.set("Authorization", `Bearer ${token}`);
    }
  }

  const response = await fetch(path, { ...options, headers });

  if (response.status === 401 && !options.skipAuth && refreshHandler) {
    const refreshed = await refreshHandler();
    if (refreshed) {
      return apiRequest<T>(path, { ...options, skipAuth: false });
    }
  }

  if (!response.ok) {
    throw await parseError(response);
  }

  if (response.status === 204) {
    return undefined as T;
  }

  return (await response.json()) as T;
}

export const api = {
  login(email: string, password: string) {
    return apiRequest<LoginResponse>("/api/auth/login", {
      method: "POST",
      body: JSON.stringify({ email, password }),
      skipAuth: true,
    });
  },
  refresh(refreshToken: string) {
    return apiRequest<RefreshResponse>("/api/auth/refresh", {
      method: "POST",
      body: JSON.stringify({ refresh_token: refreshToken }),
      skipAuth: true,
    });
  },
  logout(refreshToken: string) {
    return apiRequest<{ status: string }>("/api/auth/logout", {
      method: "POST",
      body: JSON.stringify({ refresh_token: refreshToken }),
      skipAuth: true,
    });
  },
  me() {
    return apiRequest<User>("/api/auth/me");
  },
};
