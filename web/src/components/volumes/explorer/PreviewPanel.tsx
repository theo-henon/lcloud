import { Button } from "@/components/ui/button";
import { api } from "@/lib/api";
import type { FileEntry } from "@/lib/api";
import { ActionIcon } from "@/lib/icons";

type PreviewPanelProps = {
  entry: FileEntry;
  volumeId: string;
  content: React.ReactNode;
  onClose: () => void;
};

export function PreviewPanel({ entry, volumeId, content, onClose }: PreviewPanelProps) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 p-6">
      <div className="flex max-h-[90vh] w-full max-w-5xl flex-col overflow-hidden rounded-lg border border-hairline bg-surface-card shadow-xl">
        <div className="flex items-center justify-between border-b border-hairline px-4 py-3">
          <div>
            <p className="font-semibold text-ink">{entry.name}</p>
            <p className="text-xs text-muted">{entry.path}</p>
          </div>
          <div className="flex gap-2">
            {entry.type === "file" ? (
              <Button
                variant="outline"
                onClick={() => {
                  void api.downloadFile(volumeId, entry.path, entry.name);
                }}
              >
                <ActionIcon action="download" className="mr-2" />
                Download
              </Button>
            ) : null}
            <Button variant="ghost" className="h-9 w-9 px-0" onClick={onClose} aria-label="Close">
              <ActionIcon action="close" />
            </Button>
          </div>
        </div>
        <div className="flex-1 overflow-auto p-4">{content}</div>
      </div>
    </div>
  );
}
