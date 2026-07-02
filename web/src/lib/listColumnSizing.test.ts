import { describe, expect, it } from "vitest";
import { autoFitColumnWidth, clampColumnWidth, columnCellText } from "@/lib/listColumnSizing";
import type { FileEntry } from "@/lib/api";

const sampleEntries: FileEntry[] = [
  {
    name: "short.txt",
    path: "short.txt",
    type: "file",
    size_bytes: 12,
    mime_type: "text/plain",
  },
  {
    name: "a-very-long-filename-for-testing-autofit-behavior.docx",
    path: "a-very-long-filename-for-testing-autofit-behavior.docx",
    type: "file",
    size_bytes: 4096,
    mime_type: "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
  },
  {
    name: "archive",
    path: "archive",
    type: "directory",
  },
];

describe("listColumnSizing", () => {
  it("clamps column widths", () => {
    expect(clampColumnWidth(20)).toBe(72);
    expect(clampColumnWidth(500)).toBe(500);
    expect(clampColumnWidth(5000)).toBe(960);
  });

  it("formats directory names with trailing slash for name column", () => {
    expect(columnCellText("name", sampleEntries[2])).toBe("archive/");
  });

  it("auto-fits the name column wider for long filenames", () => {
    const width = autoFitColumnWidth("name", sampleEntries);
    expect(width).toBeGreaterThan(autoFitColumnWidth("size", sampleEntries));
  });
});
