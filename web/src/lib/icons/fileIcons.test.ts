import { describe, expect, it } from "vitest";
import { getFileIcon } from "@/lib/icons/fileIcons";
import {
  File,
  FileArchive,
  FileImage,
  FileText,
  FileType,
  FileVideo,
  Folder,
} from "lucide-react";

describe("getFileIcon", () => {
  it("returns Folder for directories", () => {
    expect(getFileIcon({ type: "directory", name: "docs" })).toBe(Folder);
  });

  it("returns FileImage for image mime types", () => {
    expect(
      getFileIcon({ type: "file", name: "photo.jpg", mime_type: "image/jpeg" }),
    ).toBe(FileImage);
  });

  it("returns FileVideo for video mime types", () => {
    expect(
      getFileIcon({ type: "file", name: "clip.mp4", mime_type: "video/mp4" }),
    ).toBe(FileVideo);
  });

  it("returns FileType for pdf", () => {
    expect(getFileIcon({ type: "file", name: "report.pdf" })).toBe(FileType);
  });

  it("returns FileText for text extension fallback", () => {
    expect(getFileIcon({ type: "file", name: "notes.txt" })).toBe(FileText);
  });

  it("returns FileArchive for zip extension fallback", () => {
    expect(getFileIcon({ type: "file", name: "backup.zip" })).toBe(FileArchive);
  });

  it("returns generic File for unknown types", () => {
    expect(getFileIcon({ type: "file", name: "data.bin" })).toBe(File);
  });
});
