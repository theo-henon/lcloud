import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { ThumbnailPreview } from "@/components/volumes/ThumbnailPreview";
import type { FileEntry } from "@/lib/api";
import { formatBytes } from "@/lib/utils";

type FileBrowserProps = {
  volumeId: string;
  entries: FileEntry[];
  onOpenDirectory: (path: string) => void;
  onDownload: (path: string) => void;
  onDelete: (path: string) => void;
};

export function FileBrowser({
  volumeId,
  entries,
  onOpenDirectory,
  onDownload,
  onDelete,
}: FileBrowserProps) {
  if (entries.length === 0) {
    return (
      <Card className="p-6 text-sm text-muted">
        This folder is empty. Upload a file or create a subfolder.
      </Card>
    );
  }

  return (
    <div className="overflow-hidden rounded-lg border border-hairline">
      <table className="min-w-full divide-y divide-hairline">
        <thead className="bg-surface-soft">
          <tr>
            <th className="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-muted">
              Name
            </th>
            <th className="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-muted">
              Size
            </th>
            <th className="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-muted">
              Modified
            </th>
            <th className="px-4 py-3 text-right text-xs font-semibold uppercase tracking-wide text-muted">
              Actions
            </th>
          </tr>
        </thead>
        <tbody className="divide-y divide-hairline bg-surface-card">
          {entries.map((entry) => (
            <tr key={entry.path}>
              <td className="px-4 py-3">
                <div className="flex items-center gap-3">
                  {entry.type === "file" && entry.has_thumbnail ? (
                    <ThumbnailPreview volumeId={volumeId} path={entry.path} />
                  ) : null}
                  {entry.type === "directory" ? (
                    <button
                      type="button"
                      className="font-medium text-primary hover:underline"
                      onClick={() => onOpenDirectory(entry.path)}
                    >
                      {entry.name}/
                    </button>
                  ) : (
                    <span className="font-medium text-ink">{entry.name}</span>
                  )}
                </div>
              </td>
              <td className="px-4 py-3 text-sm text-body">
                {entry.type === "file" ? formatBytes(entry.size_bytes ?? 0) : "—"}
              </td>
              <td className="px-4 py-3 text-sm text-body">
                {entry.modified_at
                  ? new Date(entry.modified_at).toLocaleString()
                  : "—"}
              </td>
              <td className="px-4 py-3 text-right">
                {entry.type === "file" ? (
                  <div className="flex justify-end gap-2">
                    <Button variant="outline" onClick={() => onDownload(entry.path)}>
                      Download
                    </Button>
                    <Button variant="outline" onClick={() => onDelete(entry.path)}>
                      Delete
                    </Button>
                  </div>
                ) : null}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
