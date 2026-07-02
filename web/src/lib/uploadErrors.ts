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

export function formatFileOpError(error: unknown): string {
  if (error instanceof ApiError) {
    if (error.code === "PATH_EXISTS") {
      return "A file with that name already exists in the destination folder.";
    }
    if (error.code === "NOT_A_FILE") {
      return "Only files can be moved this way.";
    }
    if (error.code === "NOT_FOUND") {
      return "The file or folder could not be found.";
    }
    return error.message;
  }
  if (error instanceof Error) {
    return error.message;
  }
  return "File operation failed.";
}
