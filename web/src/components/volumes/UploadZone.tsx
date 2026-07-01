import { useRef, useState } from "react";
import { Button } from "@/components/ui/button";
import { ApiError } from "@/lib/api";

type UploadZoneProps = {
  uploading?: boolean;
  onUpload: (files: FileList) => void;
};

export function UploadZone({ uploading, onUpload }: UploadZoneProps) {
  const inputRef = useRef<HTMLInputElement>(null);
  const [dragOver, setDragOver] = useState(false);

  return (
    <div
      className={`rounded-lg border border-dashed p-6 text-center transition-colors ${
        dragOver ? "border-primary bg-surface-elevated" : "border-hairline bg-surface-soft"
      }`}
      onDragOver={(event) => {
        event.preventDefault();
        setDragOver(true);
      }}
      onDragLeave={() => setDragOver(false)}
      onDrop={(event) => {
        event.preventDefault();
        setDragOver(false);
        if (event.dataTransfer.files.length > 0) {
          onUpload(event.dataTransfer.files);
        }
      }}
    >
      <p className="text-sm text-body">Drop a file here or choose one to upload.</p>
      <Button
        type="button"
        className="mt-4"
        disabled={uploading}
        onClick={() => inputRef.current?.click()}
      >
        {uploading ? "Uploading..." : "Choose file"}
      </Button>
      <input
        ref={inputRef}
        type="file"
        className="hidden"
        onChange={(event) => {
          if (event.target.files && event.target.files.length > 0) {
            onUpload(event.target.files);
            event.target.value = "";
          }
        }}
      />
    </div>
  );
}

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
  return "Upload failed.";
}
