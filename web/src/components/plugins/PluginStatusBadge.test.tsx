import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { PluginStatusBadge } from "@/components/plugins/PluginStatusBadge";

describe("PluginStatusBadge", () => {
  it("renders running status", () => {
    render(<PluginStatusBadge status="running" />);
    expect(screen.getByText("running")).toBeInTheDocument();
  });

  it("renders error status", () => {
    render(<PluginStatusBadge status="error" />);
    expect(screen.getByText("error")).toBeInTheDocument();
  });
});
