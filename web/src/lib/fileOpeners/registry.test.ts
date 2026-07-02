import { describe, expect, it } from "vitest";
import type { FileEntry } from "@/lib/api";
import { resolveFileOpener, registerFileOpener } from "@/lib/fileOpeners/registry";

describe("file opener registry", () => {
  it("resolves highest priority opener", () => {
    registerFileOpener({
      id: "test.low",
      label: "Low",
      priority: 1,
      canOpen: (entry) => entry.name.endsWith(".test"),
      open: () => {},
    });
    registerFileOpener({
      id: "test.high",
      label: "High",
      priority: 50,
      canOpen: (entry) => entry.name.endsWith(".test"),
      open: () => {},
    });

    const entry: FileEntry = {
      name: "sample.test",
      path: "sample.test",
      type: "file",
    };
    expect(resolveFileOpener(entry)?.id).toBe("test.high");
  });
});
