import { useState } from "react";
import { useDraggable, useDroppable } from "@dnd-kit/core";
import { CSS } from "@dnd-kit/utilities";
import { ContextMenuItem } from "@/components/ui/context-menu";
import { ThumbnailPreview } from "@/components/volumes/ThumbnailPreview";
import { InlineRename } from "@/components/volumes/explorer/InlineRename";
import type { FileEntry } from "@/lib/api";
import { actionIcons, ContextMenuAction, getFileIcon } from "@/lib/icons";
import { cn } from "@/lib/utils";

const GRID_CARD_CLASS =
  "group rounded-lg border border-hairline bg-surface-card p-3 transition-all duration-150 hover:border-primary hover:bg-surface-elevated active:scale-[0.98] active:border-primary active:bg-surface-soft";

const GRID_PREVIEW_CLASS =
  "mb-3 flex h-24 items-center justify-center rounded-md bg-surface-soft transition-colors group-hover:bg-surface-card group-active:bg-surface-elevated";

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
          GRID_CARD_CLASS,
          "cursor-pointer",
          isOver && "scale-[1.02] border-primary bg-surface-elevated ring-2 ring-inset ring-primary",
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
      className={cn(
        GRID_CARD_CLASS,
        "cursor-grab active:cursor-grabbing",
        isDragging && "scale-95 opacity-50",
      )}
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
    const EmptyIcon = actionIcons.open;
    return (
      <div className="flex flex-col items-center gap-3 rounded-lg border border-hairline bg-surface-card p-8 text-center text-sm text-muted">
        <EmptyIcon className="h-8 w-8 text-muted" aria-hidden />
        <p>This folder is empty. Drop files here or create a subfolder.</p>
      </div>
    );
  }

  return (
    <>
      <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5">
        {entries.map((entry) => {
          const EntryIcon = getFileIcon(entry);
          return (
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
                <div className={GRID_PREVIEW_CLASS}>
                  {entry.type === "file" && entry.has_thumbnail ? (
                    <ThumbnailPreview volumeId={volumeId} path={entry.path} />
                  ) : (
                    <EntryIcon
                      className={cn(
                        "h-10 w-10 transition-colors duration-150",
                        entry.type === "directory"
                          ? "text-primary group-hover:scale-110"
                          : "text-muted group-hover:text-body",
                      )}
                      aria-hidden
                    />
                  )}
                </div>
                <InlineRename
                  value={entry.name}
                  active={renamingPath === entry.path}
                  onRequestRename={() => onRenameRequest(entry)}
                  onCommit={(next) => onRenameCommit(entry, next)}
                  onCancel={onRenameCancel}
                >
                  <p className="truncate text-sm font-medium text-ink transition-colors duration-150 group-hover:text-body-strong">
                    {entry.name}
                  </p>
                </InlineRename>
              </div>
            </GridCard>
          );
        })}
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
            <ContextMenuAction action="open">Open</ContextMenuAction>
          </ContextMenuItem>
          {menu.entry.type === "file" ? (
            <ContextMenuItem
              onSelect={() => {
                onDownload(menu.entry);
                setMenu(null);
              }}
            >
              <ContextMenuAction action="download">Download</ContextMenuAction>
            </ContextMenuItem>
          ) : null}
          <ContextMenuItem
            onSelect={() => {
              onRenameRequest(menu.entry);
              setMenu(null);
            }}
          >
            <ContextMenuAction action="rename">Rename</ContextMenuAction>
          </ContextMenuItem>
          {menu.entry.type === "file" ? (
            <ContextMenuItem
              onSelect={() => {
                onDelete(menu.entry);
                setMenu(null);
              }}
            >
              <ContextMenuAction action="delete" destructive>
                Delete
              </ContextMenuAction>
            </ContextMenuItem>
          ) : null}
        </div>
      ) : null}
    </>
  );
}
