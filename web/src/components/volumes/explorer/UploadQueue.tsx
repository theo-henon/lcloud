import { Button } from "@/components/ui/button";
import type { UploadQueueItem } from "@/hooks/useUploadQueue";

type UploadQueueProps = {
  items: UploadQueueItem[];
  onDismiss: (id: string) => void;
  onCancel: (id: string) => void;
};

export function UploadQueue({ items, onDismiss, onCancel }: UploadQueueProps) {
  if (items.length === 0) {
    return null;
  }

  return (
    <div className="space-y-2 rounded-lg border border-hairline bg-surface-soft p-3">
      {items.map((item) => (
        <div key={item.id} className="flex items-center justify-between gap-3 text-sm">
          <div className="min-w-0 flex-1">
            <p className="truncate font-medium text-ink">{item.file.name}</p>
            <p className="text-xs text-muted">
              {item.state === "queued" && "Queued"}
              {item.state === "uploading" && "Uploading…"}
              {item.state === "done" && "Done"}
              {item.state === "error" && (item.error ?? "Upload failed")}
            </p>
          </div>
          <div className="flex gap-2">
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
      ))}
    </div>
  );
}
