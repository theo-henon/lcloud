import { Button } from "@/components/ui/button";
import type { UploadQueueItem } from "@/hooks/useUploadQueue";
import { cn, formatBytes, formatSpeed } from "@/lib/utils";

type UploadQueueProps = {
  items: UploadQueueItem[];
  onDismiss: (id: string) => void;
  onCancel: (id: string) => void;
};

function UploadProgressBar({
  value,
  tone = "primary",
}: {
  value: number;
  tone?: "primary" | "muted" | "success" | "error";
}) {
  const fillClass =
    tone === "success"
      ? "bg-accent-emerald"
      : tone === "error"
        ? "bg-accent-rose"
        : tone === "muted"
          ? "bg-hairline-strong"
          : "bg-primary";

  return (
    <div className="mt-2 h-1.5 w-full overflow-hidden rounded-full bg-surface-elevated">
      <div
        className={cn("h-full rounded-full transition-[width] duration-150", fillClass)}
        style={{ width: `${Math.min(100, Math.max(0, value))}%` }}
      />
    </div>
  );
}

function uploadStatusText(item: UploadQueueItem): string {
  switch (item.state) {
    case "queued":
      return "Queued";
    case "uploading": {
      const percent = item.progress ?? 0;
      const speed = formatSpeed(item.speedBytesPerSec ?? 0);
      const uploaded = formatBytes(item.bytesUploaded ?? 0);
      const total = formatBytes(item.bytesTotal ?? item.file.size);
      return `Uploading · ${percent}% · ${speed} · ${uploaded} / ${total}`;
    }
    case "done":
      return "Done";
    case "error":
      return item.error ?? "Upload failed";
    default:
      return "";
  }
}

export function UploadQueue({ items, onDismiss, onCancel }: UploadQueueProps) {
  if (items.length === 0) {
    return null;
  }

  return (
    <div className="space-y-3 rounded-lg border border-hairline bg-surface-soft p-3">
      {items.map((item) => {
        const showProgress = item.state === "queued" || item.state === "uploading" || item.state === "done";
        const progressValue =
          item.state === "queued"
            ? 0
            : item.state === "done"
              ? 100
              : (item.progress ?? 0);
        const progressTone =
          item.state === "queued"
            ? "muted"
            : item.state === "done"
              ? "success"
              : item.state === "error"
                ? "error"
                : "primary";

        return (
          <div key={item.id} className="flex items-start justify-between gap-3 text-sm">
            <div className="min-w-0 flex-1">
              <p className="truncate font-medium text-ink">{item.file.name}</p>
              <p
                className={cn(
                  "text-xs",
                  item.state === "error" ? "text-accent-rose" : "text-muted",
                )}
              >
                {uploadStatusText(item)}
              </p>
              {showProgress ? (
                <UploadProgressBar value={progressValue} tone={progressTone} />
              ) : item.state === "error" ? (
                <UploadProgressBar value={item.progress ?? 0} tone="error" />
              ) : null}
            </div>
            <div className="flex shrink-0 gap-2 pt-0.5">
              {item.state === "queued" ? (
                <Button variant="ghost" onClick={() => onCancel(item.id)}>
                  Cancel
                </Button>
              ) : null}
              {item.state === "done" || item.state === "error" ? (
                <Button variant="ghost" onClick={() => onDismiss(item.id)}>
                  Dismiss
                </Button>
              ) : null}
            </div>
          </div>
        );
      })}
    </div>
  );
}
