import { useState } from "react";
import { useDraggable, useDroppable } from "@dnd-kit/core";
import { CSS } from "@dnd-kit/utilities";
import { File, Folder } from "lucide-react";
import { ContextMenuItem } from "@/components/ui/context-menu";
import { ThumbnailPreview } from "@/components/volumes/ThumbnailPreview";
import { InlineRename } from "@/components/volumes/explorer/InlineRename";
import type { FileEntry } from "@/lib/api";
import { cn } from "@/lib/utils";

type FileGridViewProps = {
  volumeId: string;
  entries: FileEntry[];
  renamingPath: string | null;
  onOpenDirectory: (path: string) => void;
  onOpenEntry: (entry: FileEntry) => void;
  onDownload: (entry: FileEntry) => void;
  onDelete: (entry: FileEntry) => void;
  onRenameRequest: (entry: FileEntry) => void;
  onRenameCommit: (entry: FileEntry, newName: string) => void;
  onRenameCancel: () => void;
};

function GridCard({
  entry,
  children,
}: {
  entry: FileEntry;
  children: React.ReactNode;
}) {
  if (entry.type === "directory") {
    const { setNodeRef, isOver } = useDroppable({
      id: `grid-folder:${entry.path}`,
      data: { type: "folder", path: entry.path },
    });
    return (
      <div
        ref={setNodeRef}
        className={cn(
          "rounded-lg border border-hairline bg-surface-card p-3",
          isOver && "ring-2 ring-primary",
        )}
      >
        {children}
      </div>
    );
  }

  const { attributes, listeners, setNodeRef, transform, isDragging } = useDraggable({
    id: `file:${entry.path}`,
    data: { type: "file", path: entry.path },
  });
  const style = transform
    ? { transform: CSS.Translate.toString(transform), opacity: isDragging ? 0.5 : 1 }
    : undefined;

  return (
    <div
      ref={setNodeRef}
      style={style}
      className="rounded-lg border border-hairline bg-surface-card p-3"
      {...listeners}
      {...attributes}
    >
      {children}
    </div>
  );
}

export function FileGridView({
  volumeId,
  entries,
  renamingPath,
  onOpenDirectory,
  onOpenEntry,
  onDownload,
  onDelete,
  onRenameRequest,
  onRenameCommit,
  onRenameCancel,
}: FileGridViewProps) {
  const [menu, setMenu] = useState<{ entry: FileEntry; x: number; y: number } | null>(null);

  if (entries.length === 0) {
    return (
      <div className="rounded-lg border border-hairline bg-surface-card p-6 text-sm text-muted">
        This folder is empty. Drop files here or create a subfolder.
      </div>
    );
  }

  return (
    <>
      <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5">
        {entries.map((entry) => (
          <GridCard key={entry.path} entry={entry}>
            <div
              onContextMenu={(event) => {
                event.preventDefault();
                setMenu({ entry, x: event.clientX, y: event.clientY });
              }}
              onDoubleClick={() => {
                if (entry.type === "directory") {
                  onOpenDirectory(entry.path);
                } else {
                  onOpenEntry(entry);
                }
              }}
            >
              <div className="mb-3 flex h-24 items-center justify-center rounded bg-surface-soft">
                {entry.type === "directory" ? (
                  <Folder className="h-10 w-10 text-primary" />
                ) : entry.has_thumbnail ? (
                  <ThumbnailPreview volumeId={volumeId} path={entry.path} />
                ) : (
                  <File className="h-10 w-10 text-muted" />
                )}
              </div>
              <InlineRename
                value={entry.name}
                active={renamingPath === entry.path}
                onRequestRename={() => onRenameRequest(entry)}
                onCommit={(next) => onRenameCommit(entry, next)}
                onCancel={onRenameCancel}
              >
                <p className="truncate text-sm font-medium text-ink">{entry.name}</p>
              </InlineRename>
            </div>
          </GridCard>
        ))}
      </div>
      {menu ? (
        <div
          className="fixed z-50 min-w-[160px] rounded-md border border-hairline bg-surface-elevated py-1 shadow-lg"
          style={{ top: menu.y, left: menu.x }}
          onMouseLeave={() => setMenu(null)}
        >
          <ContextMenuItem
            onSelect={() => {
              onOpenEntry(menu.entry);
              setMenu(null);
            }}
          >
            Open
          </ContextMenuItem>
          {menu.entry.type === "file" ? (
            <ContextMenuItem
              onSelect={() => {
                onDownload(menu.entry);
                setMenu(null);
              }}
            >
              Download
            </ContextMenuItem>
          ) : null}
          <ContextMenuItem
            onSelect={() => {
              onRenameRequest(menu.entry);
              setMenu(null);
            }}
          >
            Rename
          </ContextMenuItem>
          {menu.entry.type === "file" ? (
            <ContextMenuItem
              onSelect={() => {
                onDelete(menu.entry);
                setMenu(null);
              }}
            >
              Delete
            </ContextMenuItem>
          ) : null}
        </div>
      ) : null}
    </>
  );
}
