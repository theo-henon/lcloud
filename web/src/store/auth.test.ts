import { beforeEach, describe, expect, it } from "vitest";
import { useAuthStore } from "@/store/auth";

describe("auth store", () => {
  beforeEach(() => {
    sessionStorage.clear();
    useAuthStore.setState({
      accessToken: null,
      refreshToken: null,
      user: null,
      initialized: false,
    });
  });

  it("stores tokens on setTokens", () => {
    useAuthStore.getState().setTokens("access", "refresh");
    expect(useAuthStore.getState().accessToken).toBe("access");
    expect(sessionStorage.getItem("lcloud.access_token")).toBe("access");
  });

  it("clears tokens on logout", async () => {
    useAuthStore.getState().setTokens("access", "refresh");
    await useAuthStore.getState().logout();
    expect(useAuthStore.getState().accessToken).toBeNull();
    expect(sessionStorage.getItem("lcloud.access_token")).toBeNull();
  });
});
