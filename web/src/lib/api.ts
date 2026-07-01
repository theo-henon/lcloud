export type UserRole = "admin" | "user";

export interface User {
  id: string;
  email: string;
  role: UserRole;
  created_at: string;
}

export type FilterMode = "allow" | "block";

export interface VolumeFilters {
  mode: FilterMode | "";
  extensions: string[];
}

export interface DiskInfo {
  path: string;
  name: string;
  label: string;
  total_bytes: number;
  free_bytes: number;
}

export interface Volume {
  id: string;
  name: string;
  owner_id: string;
  disk_path: string;
  root_path: string;
  quota_bytes: number;
  used_bytes: number;
  filters: VolumeFilters;
  created_at: string;
  updated_at: string;
}

export interface FileEntry {
  name: string;
  path: string;
  type: "file" | "directory";
  size_bytes?: number;
  mime_type?: string;
  modified_at?: string;
  has_thumbnail?: boolean;
}

export interface DirectoryListing {
  path: string;
  entries: FileEntry[];
}

export interface CreateVolumeInput {
  name: string;
  disk_path: string;
  quota_bytes: number;
  filters: VolumeFilters;
}

export interface PatchVolumeInput {
  name?: string;
  quota_bytes?: number;
  filters?: VolumeFilters;
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
  listDisks() {
    return apiRequest<{ disks: DiskInfo[] }>("/api/disks");
  },
  listVolumes() {
    return apiRequest<{ volumes: Volume[] }>("/api/volumes");
  },
  getVolume(id: string) {
    return apiRequest<Volume>(`/api/volumes/${id}`);
  },
  createVolume(input: CreateVolumeInput) {
    return apiRequest<Volume>("/api/volumes", {
      method: "POST",
      body: JSON.stringify(input),
    });
  },
  patchVolume(id: string, input: PatchVolumeInput) {
    return apiRequest<Volume>(`/api/volumes/${id}`, {
      method: "PATCH",
      body: JSON.stringify(input),
    });
  },
  deleteVolume(id: string, force = false) {
    const query = force ? "?force=true" : "";
    return apiRequest<void>(`/api/volumes/${id}${query}`, { method: "DELETE" });
  },
  listFiles(volumeId: string, path = ".") {
    const params = new URLSearchParams({ path });
    return apiRequest<DirectoryListing>(`/api/volumes/${volumeId}/files?${params}`);
  },
  createDirectory(volumeId: string, path: string) {
    return apiRequest<{ path: string }>(`/api/volumes/${volumeId}/files/directories`, {
      method: "POST",
      body: JSON.stringify({ path }),
    });
  },
  async uploadFile(volumeId: string, file: File, path = ".") {
    const formData = new FormData();
    formData.append("file", file);
    formData.append("path", path);

    const headers = new Headers();
    if (accessTokenProvider) {
      const token = accessTokenProvider();
      if (token) {
        headers.set("Authorization", `Bearer ${token}`);
      }
    }

    const response = await fetch(`/api/volumes/${volumeId}/files`, {
      method: "POST",
      headers,
      body: formData,
    });

    if (response.status === 401 && refreshHandler) {
      const refreshed = await refreshHandler();
      if (refreshed) {
        return api.uploadFile(volumeId, file, path);
      }
    }

    if (!response.ok) {
      throw await parseError(response);
    }
    return (await response.json()) as FileEntry;
  },
  fileContentUrl(volumeId: string, path: string) {
    const params = new URLSearchParams({ path });
    return `/api/volumes/${volumeId}/files/content?${params}`;
  },
  fileThumbnailUrl(volumeId: string, path: string) {
    const params = new URLSearchParams({ path });
    return `/api/volumes/${volumeId}/files/thumbnail?${params}`;
  },
  deleteFile(volumeId: string, path: string) {
    const params = new URLSearchParams({ path });
    return apiRequest<void>(`/api/volumes/${volumeId}/files?${params}`, { method: "DELETE" });
  },
  async downloadFile(volumeId: string, path: string, filename: string) {
    const headers = new Headers();
    if (accessTokenProvider) {
      const token = accessTokenProvider();
      if (token) {
        headers.set("Authorization", `Bearer ${token}`);
      }
    }

    const response = await fetch(api.fileContentUrl(volumeId, path), { headers });
    if (!response.ok) {
      throw await parseError(response);
    }
    const blob = await response.blob();
    const url = URL.createObjectURL(blob);
    const anchor = document.createElement("a");
    anchor.href = url;
    anchor.download = filename;
    anchor.click();
    URL.revokeObjectURL(url);
  },
};
