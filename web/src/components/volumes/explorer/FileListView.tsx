import { useState } from "react";
import { useDraggable, useDroppable } from "@dnd-kit/core";
import { CSS } from "@dnd-kit/utilities";
import { File, Folder } from "lucide-react";
import { ContextMenuItem } from "@/components/ui/context-menu";
import { ThumbnailPreview } from "@/components/volumes/ThumbnailPreview";
import { InlineRename } from "@/components/volumes/explorer/InlineRename";
import { ListColumnHeader } from "@/components/volumes/explorer/ListColumnPicker";
import type { ColumnId, ListColumnPref } from "@/hooks/useExplorerPrefs";
import { visibleListColumns } from "@/hooks/useExplorerPrefs";
import type { FileEntry } from "@/lib/api";
import { cn, formatBytes } from "@/lib/utils";

type FileListViewProps = {
  volumeId: string;
  entries: FileEntry[];
  listColumns: ListColumnPref[];
  onColumnToggle: (id: ColumnId, visible: boolean) => void;
  renamingPath: string | null;
  onOpenDirectory: (path: string) => void;
  onOpenEntry: (entry: FileEntry) => void;
  onDownload: (entry: FileEntry) => void;
  onDelete: (entry: FileEntry) => void;
  onRenameRequest: (entry: FileEntry) => void;
  onRenameCommit: (entry: FileEntry, newName: string) => void;
  onRenameCancel: () => void;
};

function DirectoryRow({
  entry,
  children,
}: {
  entry: FileEntry;
  children: React.ReactNode;
}) {
  const { setNodeRef, isOver } = useDroppable({
    id: `list-folder:${entry.path}`,
    data: { type: "folder", path: entry.path },
  });
  return (
    <tr
      ref={setNodeRef}
      className={cn("hover:bg-surface-elevated", isOver && "bg-surface-elevated ring-1 ring-inset ring-primary")}
    >
      {children}
    </tr>
  );
}

function FileRow({
  entry,
  children,
}: {
  entry: FileEntry;
  children: React.ReactNode;
}) {
  const { attributes, listeners, setNodeRef, transform, isDragging } = useDraggable({
    id: `file:${entry.path}`,
    data: { type: "file", path: entry.path },
  });
  const style = transform
    ? { transform: CSS.Translate.toString(transform), opacity: isDragging ? 0.5 : 1 }
    : undefined;
  return (
    <tr
      ref={setNodeRef}
      style={style}
      className="hover:bg-surface-elevated"
      {...listeners}
      {...attributes}
    >
      {children}
    </tr>
  );
}

function RowContextMenu({
  entry,
  onOpenEntry,
  onDownload,
  onDelete,
  onRenameRequest,
  children,
}: {
  entry: FileEntry;
  onOpenEntry: (entry: FileEntry) => void;
  onDownload: (entry: FileEntry) => void;
  onDelete: (entry: FileEntry) => void;
  onRenameRequest: (entry: FileEntry) => void;
  children: (props: { onContextMenu: (event: React.MouseEvent) => void }) => React.ReactNode;
}) {
  const [menu, setMenu] = useState<{ x: number; y: number } | null>(null);
  return (
    <>
      {children({
        onContextMenu: (event) => {
          event.preventDefault();
          setMenu({ x: event.clientX, y: event.clientY });
        },
      })}
      {menu ? (
        <div
          className="fixed z-50 min-w-[160px] rounded-md border border-hairline bg-surface-elevated py-1 shadow-lg"
          style={{ top: menu.y, left: menu.x }}
          onMouseLeave={() => setMenu(null)}
        >
          <ContextMenuItem onSelect={() => { onOpenEntry(entry); setMenu(null); }}>Open</ContextMenuItem>
          {entry.type === "file" ? (
            <ContextMenuItem onSelect={() => { onDownload(entry); setMenu(null); }}>Download</ContextMenuItem>
          ) : null}
          <ContextMenuItem onSelect={() => { onRenameRequest(entry); setMenu(null); }}>Rename</ContextMenuItem>
          {entry.type === "file" ? (
            <ContextMenuItem onSelect={() => { onDelete(entry); setMenu(null); }}>Delete</ContextMenuItem>
          ) : null}
        </div>
      ) : null}
    </>
  );
}

function renderCell(
  column: ListColumnPref,
  entry: FileEntry,
  volumeId: string,
  renamingPath: string | null,
  onOpenDirectory: (path: string) => void,
  onRenameRequest: (entry: FileEntry) => void,
  onRenameCommit: (entry: FileEntry, newName: string) => void,
  onRenameCancel: () => void,
) {
  switch (column.id) {
    case "name":
      return (
        <InlineRename
          value={entry.name}
          active={renamingPath === entry.path}
          onRequestRename={() => onRenameRequest(entry)}
          onCommit={(next) => onRenameCommit(entry, next)}
          onCancel={onRenameCancel}
        >
          <div className="flex items-center gap-3">
            {entry.type === "file" && entry.has_thumbnail ? (
              <ThumbnailPreview volumeId={volumeId} path={entry.path} />
            ) : entry.type === "directory" ? (
              <Folder className="h-5 w-5 text-primary" />
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
        </InlineRename>
      );
    case "size":
      return entry.type === "file" ? formatBytes(entry.size_bytes ?? 0) : "—";
    case "modified":
      return entry.modified_at ? new Date(entry.modified_at).toLocaleString() : "—";
    case "type":
      return entry.mime_type ?? (entry.type === "directory" ? "Folder" : "—");
    default:
      return "—";
  }
}

export function FileListView(props: FileListViewProps) {
  const {
    volumeId,
    entries,
    listColumns,
    onColumnToggle,
    renamingPath,
    onOpenDirectory,
    onOpenEntry,
    onDownload,
    onDelete,
    onRenameRequest,
    onRenameCommit,
    onRenameCancel,
  } = props;

  const columns = visibleListColumns(listColumns);

  if (entries.length === 0) {
    return (
      <div className="rounded-lg border border-hairline bg-surface-card p-6 text-sm text-muted">
        This folder is empty. Drop files here or create a subfolder.
      </div>
    );
  }

  return (
    <div className="overflow-hidden rounded-lg border border-hairline">
      <table className="min-w-full divide-y divide-hairline">
        <ListColumnHeader listColumns={listColumns} onColumnToggle={onColumnToggle} />
          <tbody className="divide-y divide-hairline bg-surface-card">
          {entries.map((entry) => (
            <RowContextMenu
              key={entry.path}
              entry={entry}
              onOpenEntry={onOpenEntry}
              onDownload={onDownload}
              onDelete={onDelete}
              onRenameRequest={onRenameRequest}
            >
              {({ onContextMenu }) => {
                const Row = entry.type === "directory" ? DirectoryRow : FileRow;
                return (
                  <Row entry={entry}>
                    {columns.map((column) => (
                      <td
                        key={column.id}
                        className="px-4 py-3 text-sm text-body"
                        onContextMenu={onContextMenu}
                        onDoubleClick={() => {
                          if (entry.type === "directory") {
                            onOpenDirectory(entry.path);
                          } else {
                            onOpenEntry(entry);
                          }
                        }}
                      >
                        {renderCell(
                          column,
                          entry,
                          volumeId,
                          renamingPath,
                          onOpenDirectory,
                          onRenameRequest,
                          onRenameCommit,
                          onRenameCancel,
                        )}
                      </td>
                    ))}
                    <td className="w-10" aria-hidden />
                  </Row>
                );
              }}
            </RowContextMenu>
          ))}
        </tbody>
      </table>
    </div>
  );
}
