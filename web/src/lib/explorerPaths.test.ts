import { describe, expect, it } from "vitest";
import { treeExpandedPathsForSelection } from "@/lib/explorerPaths";

describe("treeExpandedPathsForSelection", () => {
  it("expands only root at the volume root", () => {
    expect(treeExpandedPathsForSelection(".")).toEqual(["."]);
  });

  it("expands root and the selected folder for a single segment path", () => {
    expect(treeExpandedPathsForSelection("test")).toEqual([".", "test"]);
  });

  it("expands each ancestor through the selected nested folder", () => {
    expect(treeExpandedPathsForSelection("test/nested")).toEqual([".", "test", "test/nested"]);
  });
});
