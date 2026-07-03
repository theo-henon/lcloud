import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it, vi } from "vitest";
import { QuickActionsWidget } from "@/components/dashboard/widgets/QuickActionsWidget";

vi.mock("@/store/auth", () => ({
  useAuthStore: (selector: (state: { user: { email: string; role: string } }) => unknown) =>
    selector({ user: { email: "user@example.com", role: "user" } }),
}));

describe("QuickActionsWidget", () => {
  it("renders navigation links", () => {
    render(
      <MemoryRouter>
        <QuickActionsWidget id="qa-1" type="quick-actions" />
      </MemoryRouter>,
    );

    expect(screen.getByRole("link", { name: "New volume" })).toHaveAttribute("href", "/volumes");
    expect(screen.getByRole("link", { name: "Monitoring" })).toHaveAttribute("href", "/monitoring");
    expect(screen.getByRole("link", { name: "Tasks" })).toHaveAttribute("href", "/tasks");
    expect(screen.getByRole("link", { name: "Plugins" })).toHaveAttribute("href", "/plugins");
  });
});
