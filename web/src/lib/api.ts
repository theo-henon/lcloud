export type UserRole = "admin" | "user";

export type CreateAdminUserInput = {
  email: string;
  password: string;
  role?: UserRole;
};

export type PatchAdminUserInput = {
  role?: UserRole;
  disabled?: boolean;
  password?: string;
};

export type PluginStatus = "running" | "stopped" | "error";

export interface PluginRecord {
  id: string;
  name: string;
  version: string;
  description: string;
  enabled: boolean;
  status: PluginStatus;
  subscriptions: string[];
  last_error: string;
  discovered_at: string;
}

export interface PluginLogEntry {
  id: string;
  plugin_id: string;
  volume_id?: string;
  event_type: string;
  level: string;
  message: string;
  created_at: string;
}

export interface User {
  id: string;
  email: string;
  role: UserRole;
  created_at: string;
  disabled_at: string | null;
}

export type FilterMode = "allow" | "block";

export interface VolumeFilters {
  mode: FilterMode | "";
  extensions: string[];
}

export interface DiskInfo {
  id: string;
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
  owner_email?: string;
  disk_path: string;
  root_path: string;
  quota_bytes: number;
  used_bytes: number;
  filters: VolumeFilters;
  created_at: string;
  updated_at: string;
  deletion_request_pending?: boolean;
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
  disk_path?: string;
  disk_id?: string;
  quota_bytes: number;
  filters: VolumeFilters;
}

export interface VolumeDeletionRequest {
  id: string;
  volume_id: string;
  volume_name: string;
  user_id: string;
  user_email: string;
  created_at: string;
}

export interface PatchVolumeInput {
  name?: string;
  quota_bytes?: number;
  filters?: VolumeFilters;
}

export interface CategoryBreakdown {
  category: string;
  file_count: number;
  bytes: number;
  proportion: number;
}

export interface MimeBreakdown {
  mime_type: string;
  file_count: number;
  bytes: number;
  proportion: number;
}

export interface VolumeSummary {
  id: string;
  name: string;
  owner_email?: string;
  disk_path: string;
  quota_bytes: number;
  used_bytes: number;
  file_count: number;
  top_category?: string;
  stats_computed_at?: string;
}

export interface DiskOverview {
  path: string;
  name: string;
  label: string;
  total_bytes: number;
  free_bytes: number;
  used_by_volumes_bytes: number;
  volume_count: number;
}

export interface MonitoringOverview {
  disks: DiskOverview[];
  volumes: VolumeSummary[];
}

export interface VolumeStats {
  volume_id: string;
  name: string;
  quota_bytes: number;
  used_bytes: number;
  free_bytes?: number;
  usage_percent?: number;
  file_count: number;
  computed_at: string;
  by_category: CategoryBreakdown[];
  by_mime: MimeBreakdown[];
}

export interface SearchResultItem {
  name: string;
  relative_path: string;
  mime_type: string;
  size_bytes: number;
  modified_at: string;
  has_thumbnail: boolean;
}

export interface GlobalSearchResultItem extends SearchResultItem {
  volume_id: string;
  volume_name: string;
}

export interface SearchParams {
  q?: string;
  mime_prefix?: string;
  min_size?: number;
  max_size?: number;
  modified_after?: string;
  modified_before?: string;
  limit?: number;
}

export interface SearchResponse {
  query: Record<string, unknown>;
  total: number;
  results: SearchResultItem[];
}

export interface GlobalSearchResponse {
  query: Record<string, unknown>;
  total: number;
  results: GlobalSearchResultItem[];
}

export interface InstanceSettings {
  mask_disk_names: boolean;
  max_upload_bytes: number;
}

export type TaskScope = "volume" | "global";
export type TaskScheduleType = "cron" | "interval";
export type TaskRunStatus = "success" | "failed" | "skipped" | "dry_run";

export interface TaskRecord {
  id: string;
  owner_id: string;
  owner_email?: string;
  name: string;
  macro: string;
  scope: TaskScope;
  volume_id?: string;
  volume_name?: string;
  parameters: Record<string, unknown>;
  schedule_type: TaskScheduleType;
  schedule: string;
  schedule_description: string;
  enabled: boolean;
  next_run_at?: string;
  last_run_at?: string;
  last_run_status?: TaskRunStatus;
  created_at: string;
  updated_at: string;
}

export interface TaskRunRecord {
  id: string;
  task_id: string;
  status: TaskRunStatus;
  affected_count: number;
  message: string;
  error?: string;
  started_at: string;
  finished_at: string;
  duration_ms: number;
}

export interface CreateTaskInput {
  name: string;
  macro: string;
  scope: TaskScope;
  volume_id?: string;
  parameters?: Record<string, unknown>;
  schedule_type: TaskScheduleType;
  schedule: string;
  enabled?: boolean;
}

export interface PatchTaskInput {
  name?: string;
  macro?: string;
  scope?: TaskScope;
  volume_id?: string;
  parameters?: Record<string, unknown>;
  schedule_type?: TaskScheduleType;
  schedule?: string;
  enabled?: boolean;
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

export type UploadProgress = {
  loaded: number;
  total: number;
  percent: number;
};

async function uploadFileXHR(
  volumeId: string,
  file: File,
  path: string,
  onProgress?: (progress: UploadProgress) => void,
): Promise<FileEntry> {
  return new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest();
    const formData = new FormData();
    formData.append("file", file);
    formData.append("path", path);

    xhr.upload.addEventListener("progress", (event) => {
      if (!onProgress) {
        return;
      }
      const total = event.lengthComputable ? event.total : file.size;
      const loaded = event.loaded;
      const percent = total > 0 ? Math.min(100, Math.round((loaded / total) * 100)) : 0;
      onProgress({ loaded, total, percent });
    });

    xhr.addEventListener("load", () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        resolve(JSON.parse(xhr.responseText) as FileEntry);
        return;
      }
      try {
        const body = JSON.parse(xhr.responseText) as { error?: string; code?: string };
        reject(new ApiError(body.error ?? "Request failed", xhr.status, body.code));
      } catch {
        reject(new ApiError("Request failed", xhr.status));
      }
    });

    xhr.addEventListener("error", () => reject(new ApiError("Request failed", 0)));
    xhr.addEventListener("abort", () => reject(new ApiError("Upload cancelled", 0)));

    xhr.open("POST", `/api/volumes/${volumeId}/files`);
    if (accessTokenProvider) {
      const token = accessTokenProvider();
      if (token) {
        xhr.setRequestHeader("Authorization", `Bearer ${token}`);
      }
    }
    xhr.send(formData);
  });
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
  listAdminUsers() {
    return apiRequest<{ users: User[] }>("/api/admin/users");
  },
  createAdminUser(input: CreateAdminUserInput) {
    return apiRequest<User>("/api/admin/users", {
      method: "POST",
      body: JSON.stringify(input),
    });
  },
  patchAdminUser(id: string, input: PatchAdminUserInput) {
    return apiRequest<User>(`/api/admin/users/${id}`, {
      method: "PATCH",
      body: JSON.stringify(input),
    });
  },
  listDisks() {
    return apiRequest<{ disks: DiskInfo[] }>("/api/disks");
  },
  listVolumes(ownerId?: string) {
    const params = new URLSearchParams();
    if (ownerId) {
      params.set("owner_id", ownerId);
    }
    const query = params.toString() ? `?${params.toString()}` : "";
    return apiRequest<{ volumes: Volume[] }>(`/api/volumes${query}`);
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
  requestVolumeDeletion(id: string) {
    return apiRequest<{ status: string }>(`/api/volumes/${id}/deletion-request`, {
      method: "POST",
    });
  },
  listVolumeDeletionRequests() {
    return apiRequest<{ requests: VolumeDeletionRequest[] }>(
      "/api/admin/volume-deletion-requests",
    );
  },
  dismissVolumeDeletionRequest(id: string) {
    return apiRequest<void>(`/api/admin/volume-deletion-requests/${id}`, {
      method: "DELETE",
    });
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
  async uploadFile(
    volumeId: string,
    file: File,
    path = ".",
    onProgress?: (progress: UploadProgress) => void,
  ) {
    try {
      return await uploadFileXHR(volumeId, file, path, onProgress);
    } catch (error) {
      if (error instanceof ApiError && error.status === 401 && refreshHandler) {
        const refreshed = await refreshHandler();
        if (refreshed) {
          return uploadFileXHR(volumeId, file, path, onProgress);
        }
      }
      throw error;
    }
  },
  fileContentUrl(volumeId: string, path: string, disposition: "inline" | "attachment" = "attachment") {
    const params = new URLSearchParams({ path, disposition });
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
  moveFile(volumeId: string, fromPath: string, toPath: string) {
    return apiRequest<FileEntry>(`/api/volumes/${volumeId}/files/move`, {
      method: "PATCH",
      body: JSON.stringify({ from_path: fromPath, to_path: toPath }),
    });
  },
  renameFile(volumeId: string, path: string, newName: string) {
    return apiRequest<FileEntry>(`/api/volumes/${volumeId}/files/rename`, {
      method: "PATCH",
      body: JSON.stringify({ path, new_name: newName }),
    });
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
  getMonitoringOverview() {
    return apiRequest<MonitoringOverview>("/api/monitoring/overview");
  },
  getVolumeStats(volumeId: string) {
    return apiRequest<VolumeStats>(`/api/monitoring/volumes/${volumeId}/stats`);
  },
  refreshVolumeStats(volumeId: string) {
    return apiRequest<VolumeStats>(`/api/monitoring/volumes/${volumeId}/stats/refresh`, {
      method: "POST",
    });
  },
  searchVolumeFiles(volumeId: string, params: SearchParams = {}) {
    const query = buildSearchQuery(params);
    return apiRequest<SearchResponse>(
      `/api/volumes/${volumeId}/search${query ? `?${query}` : ""}`,
    );
  },
  searchAllFiles(params: SearchParams = {}) {
    const query = buildSearchQuery(params);
    return apiRequest<GlobalSearchResponse>(`/api/search${query ? `?${query}` : ""}`);
  },
  getSettings() {
    return apiRequest<InstanceSettings>("/api/settings");
  },
  patchAdminSettings(input: { mask_disk_names: boolean }) {
    return apiRequest<InstanceSettings>("/api/admin/settings", {
      method: "PATCH",
      body: JSON.stringify(input),
    });
  },
  getPlugins() {
    return apiRequest<{ plugins: PluginRecord[] }>("/api/plugins");
  },
  getPluginLogs(params: { plugin_id?: string; volume_id?: string; limit?: number } = {}) {
    const search = new URLSearchParams();
    if (params.plugin_id) search.set("plugin_id", params.plugin_id);
    if (params.volume_id) search.set("volume_id", params.volume_id);
    if (params.limit != null) search.set("limit", String(params.limit));
    const query = search.toString();
    return apiRequest<{ entries: PluginLogEntry[] }>(
      `/api/plugins/logs${query ? `?${query}` : ""}`,
    );
  },
  patchPlugin(id: string, input: { enabled: boolean }) {
    return apiRequest<PluginRecord>(`/api/plugins/${id}`, {
      method: "PATCH",
      body: JSON.stringify(input),
    });
  },
  listTasks(filters?: { volumeId?: string; ownerId?: string }) {
    const params = new URLSearchParams();
    if (filters?.volumeId) {
      params.set("volume_id", filters.volumeId);
    }
    if (filters?.ownerId) {
      params.set("owner_id", filters.ownerId);
    }
    const query = params.toString() ? `?${params.toString()}` : "";
    return apiRequest<{ tasks: TaskRecord[] }>(`/api/tasks${query}`);
  },
  getTask(id: string) {
    return apiRequest<TaskRecord>(`/api/tasks/${id}`);
  },
  createTask(input: CreateTaskInput) {
    return apiRequest<TaskRecord>("/api/tasks", {
      method: "POST",
      body: JSON.stringify(input),
    });
  },
  patchTask(id: string, input: PatchTaskInput) {
    return apiRequest<TaskRecord>(`/api/tasks/${id}`, {
      method: "PATCH",
      body: JSON.stringify(input),
    });
  },
  deleteTask(id: string) {
    return apiRequest<{ status: string }>(`/api/tasks/${id}`, { method: "DELETE" });
  },
  runTask(id: string, dryRun = false) {
    const query = dryRun ? "?dry_run=true" : "";
    return apiRequest<{ run_id: string }>(`/api/tasks/${id}/run${query}`, {
      method: "POST",
    });
  },
  listTaskRuns(id: string, limit = 20) {
    return apiRequest<{ runs: TaskRunRecord[] }>(
      `/api/tasks/${id}/runs?limit=${limit}`,
    );
  },
};

function buildSearchQuery(params: SearchParams): string {
  const search = new URLSearchParams();
  if (params.q) search.set("q", params.q);
  if (params.mime_prefix) search.set("mime_prefix", params.mime_prefix);
  if (params.min_size != null) search.set("min_size", String(params.min_size));
  if (params.max_size != null) search.set("max_size", String(params.max_size));
  if (params.modified_after) search.set("modified_after", params.modified_after);
  if (params.modified_before) search.set("modified_before", params.modified_before);
  if (params.limit != null) search.set("limit", String(params.limit));
  return search.toString();
}
