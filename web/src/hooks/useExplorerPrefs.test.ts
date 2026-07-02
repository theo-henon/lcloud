import { describe, expect, it } from "vitest";
import { DEFAULT_EXPLORER_PREFS, loadExplorerPrefs } from "@/hooks/useExplorerPrefs";

describe("useExplorerPrefs", () => {
  it("returns defaults when storage is unavailable", () => {
    expect(loadExplorerPrefs()).toEqual(DEFAULT_EXPLORER_PREFS);
  });
});
