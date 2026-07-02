import { describe, expect, it } from "vitest";
import { ApiError } from "@/lib/api";
import { formatUploadError } from "@/lib/uploadErrors";

describe("formatUploadError", () => {
  it("includes the configured upload limit for UPLOAD_TOO_LARGE", () => {
    const error = new ApiError("upload exceeds max size", 413, "UPLOAD_TOO_LARGE");
    expect(formatUploadError(error, 104857600)).toBe(
      "This file exceeds the maximum upload size (100.0 MB).",
    );
  });
});
