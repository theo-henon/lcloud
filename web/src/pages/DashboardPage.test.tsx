import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it, vi } from "vitest";
import { DashboardPage } from "@/pages/DashboardPage";
import type { DashboardResponse } from "@/lib/dashboard/types";

const mockLayout: DashboardResponse = {
  layout: {
    version: 1,
    widgets: [
      { id: "w1", type: "welcome", x: 0, y: 0, w: 4, h: 1 },
      { id: "w2", type: "quick-actions", x: 4, y: 0, w: 4, h: 1 },
    ],
  },
  catalog: [
    {
      type: "welcome",
      title: "Welcome",
      description: "Personal greeting",
      default_w: 4,
      default_h: 1,
      admin_only: false,
    },
    {
      type: "quick-actions",
      title: "Quick actions",
      description: "Shortcuts",
      default_w: 4,
      default_h: 1,
      admin_only: false,
    },
  ],
  updated_at: null,
};

vi.mock("@/hooks/useDashboardLayout", () => ({
  useDashboardLayout: () => ({
    isLoading: false,
    isError: false,
    data: mockLayout,
  }),
  useSaveDashboardLayout: () => ({
    mutateAsync: vi.fn(async () => mockLayout),
    isPending: false,
  }),
  useResetDashboardLayout: () => ({
    mutateAsync: vi.fn(async () => undefined),
    isPending: false,
  }),
}));

vi.mock("@/store/auth", () => ({
  useAuthStore: (selector: (state: { user: { email: string; role: string } }) => unknown) =>
    selector({ user: { email: "user@example.com", role: "user" } }),
}));

function renderDashboard() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter>
        <DashboardPage />
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe("DashboardPage", () => {
  it("renders widgets instead of the Phase 0 placeholder", async () => {
    renderDashboard();

    expect(screen.queryByText(/Phase 0/i)).not.toBeInTheDocument();
    expect(await screen.findByRole("heading", { name: "Welcome" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Quick actions" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Customize" })).toBeInTheDocument();
    expect(screen.getByText("Your personal cockpit.")).toBeInTheDocument();
  });
});
