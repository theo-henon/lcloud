import { ApiError } from "@/lib/api";

export function formatUploadError(error: unknown): string {
  if (error instanceof ApiError) {
    if (error.code === "FILE_FILTER_REJECTED") {
      return "This file type is not allowed for this volume.";
    }
    if (error.code === "QUOTA_EXCEEDED") {
      return "Upload would exceed the volume quota.";
    }
    return error.message;
  }
  if (error instanceof Error) {
    return error.message;
  }
  return "Upload failed.";
}
