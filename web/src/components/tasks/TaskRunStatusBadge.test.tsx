import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { TaskRunStatusBadge } from "@/components/tasks/TaskRunStatusBadge";

describe("TaskRunStatusBadge", () => {
  it("renders dry_run label", () => {
    render(<TaskRunStatusBadge status="dry_run" />);
    expect(screen.getByText("Simulation")).toBeInTheDocument();
  });
});
