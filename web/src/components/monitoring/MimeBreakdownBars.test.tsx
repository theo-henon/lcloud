import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MimeBreakdownBars } from "@/components/monitoring/MimeBreakdownBars";

describe("MimeBreakdownBars", () => {
  it("renders category bars", () => {
    render(
      <MimeBreakdownBars
        items={[
          { category: "images", file_count: 2, bytes: 900, proportion: 0.9 },
          { category: "other", file_count: 1, bytes: 100, proportion: 0.1 },
        ]}
      />,
    );

    expect(screen.getByText("Images")).toBeInTheDocument();
    expect(screen.getByText("Other")).toBeInTheDocument();
  });
});
