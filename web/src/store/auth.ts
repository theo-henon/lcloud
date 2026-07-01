import { create } from "zustand";
import { api, configureApiAuth, type User } from "@/lib/api";

const ACCESS_TOKEN_KEY = "lcloud.access_token";
const REFRESH_TOKEN_KEY = "lcloud.refresh_token";

interface AuthState {
  accessToken: string | null;
  refreshToken: string | null;
  user: User | null;
  initialized: boolean;
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  hydrate: () => Promise<void>;
  setTokens: (accessToken: string, refreshToken: string, user?: User | null) => void;
}

export const useAuthStore = create<AuthState>((set, get) => ({
  accessToken: null,
  refreshToken: null,
  user: null,
  initialized: false,

  setTokens(accessToken, refreshToken, user = null) {
    sessionStorage.setItem(ACCESS_TOKEN_KEY, accessToken);
    sessionStorage.setItem(REFRESH_TOKEN_KEY, refreshToken);
    set({ accessToken, refreshToken, user });
  },

  async login(email, password) {
    const result = await api.login(email, password);
    get().setTokens(result.access_token, result.refresh_token, result.user);
  },

  async logout() {
    const refreshToken = get().refreshToken ?? sessionStorage.getItem(REFRESH_TOKEN_KEY);
    if (refreshToken) {
      try {
        await api.logout(refreshToken);
      } catch {
        // Ignore logout errors and clear local session anyway.
      }
    }
    sessionStorage.removeItem(ACCESS_TOKEN_KEY);
    sessionStorage.removeItem(REFRESH_TOKEN_KEY);
    set({ accessToken: null, refreshToken: null, user: null });
  },

  async hydrate() {
    const accessToken = sessionStorage.getItem(ACCESS_TOKEN_KEY);
    const refreshToken = sessionStorage.getItem(REFRESH_TOKEN_KEY);

    if (!accessToken || !refreshToken) {
      set({ initialized: true, accessToken: null, refreshToken: null, user: null });
      return;
    }

    set({ accessToken, refreshToken });

    try {
      const user = await api.me();
      set({ user, initialized: true });
    } catch {
      try {
        const refreshed = await api.refresh(refreshToken);
        get().setTokens(refreshed.access_token, refreshed.refresh_token);
        const user = await api.me();
        set({ user, initialized: true });
      } catch {
        sessionStorage.removeItem(ACCESS_TOKEN_KEY);
        sessionStorage.removeItem(REFRESH_TOKEN_KEY);
        set({
          accessToken: null,
          refreshToken: null,
          user: null,
          initialized: true,
        });
      }
    }
  },
}));

configureApiAuth(
  () => useAuthStore.getState().accessToken,
  async () => {
    const refreshToken =
      useAuthStore.getState().refreshToken ??
      sessionStorage.getItem(REFRESH_TOKEN_KEY);
    if (!refreshToken) {
      return false;
    }

    try {
      const refreshed = await api.refresh(refreshToken);
      useAuthStore
        .getState()
        .setTokens(refreshed.access_token, refreshed.refresh_token);
      return true;
    } catch {
      await useAuthStore.getState().logout();
      return false;
    }
  },
);

export function isAuthenticated() {
  return Boolean(useAuthStore.getState().accessToken);
}
