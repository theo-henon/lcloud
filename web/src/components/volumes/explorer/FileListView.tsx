import { useState } from "react";
import { useDraggable, useDroppable } from "@dnd-kit/core";
import { CSS } from "@dnd-kit/utilities";
import { ContextMenuItem } from "@/components/ui/context-menu";
import { ThumbnailPreview } from "@/components/volumes/ThumbnailPreview";
import { InlineRename } from "@/components/volumes/explorer/InlineRename";
import { ListColumnHeader } from "@/components/volumes/explorer/ListColumnHeader";
import type { ColumnId, ListColumnPref } from "@/hooks/useExplorerPrefs";
import { visibleListColumns } from "@/hooks/useExplorerPrefs";
import type { FileEntry } from "@/lib/api";
import { actionIcons, ContextMenuAction, getFileIcon } from "@/lib/icons";
import { cn, formatBytes } from "@/lib/utils";

type FileListViewProps = {
  volumeId: string;
  entries: FileEntry[];
  listColumns: ListColumnPref[];
  onColumnToggle: (id: ColumnId, visible: boolean) => void;
  onColumnWidthChange: (id: ColumnId, width: number) => void;
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
          <ContextMenuItem onSelect={() => { onOpenEntry(entry); setMenu(null); }}>
            <ContextMenuAction action="open">Open</ContextMenuAction>
          </ContextMenuItem>
          {entry.type === "file" ? (
            <ContextMenuItem onSelect={() => { onDownload(entry); setMenu(null); }}>
              <ContextMenuAction action="download">Download</ContextMenuAction>
            </ContextMenuItem>
          ) : null}
          <ContextMenuItem onSelect={() => { onRenameRequest(entry); setMenu(null); }}>
            <ContextMenuAction action="rename">Rename</ContextMenuAction>
          </ContextMenuItem>
          {entry.type === "file" ? (
            <ContextMenuItem onSelect={() => { onDelete(entry); setMenu(null); }}>
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
          <div className="flex min-w-0 items-center gap-3">
            {entry.type === "file" && entry.has_thumbnail ? (
              <ThumbnailPreview volumeId={volumeId} path={entry.path} />
            ) : (() => {
              const EntryIcon = getFileIcon(entry);
              return (
                <EntryIcon
                  className={cn(
                    "h-5 w-5 shrink-0",
                    entry.type === "directory" ? "text-primary" : "text-muted",
                  )}
                  aria-hidden
                />
              );
            })()}
            {entry.type === "directory" ? (
              <button
                type="button"
                className="truncate font-medium text-primary hover:underline"
                onClick={() => onOpenDirectory(entry.path)}
              >
                {entry.name}/
              </button>
            ) : (
              <span className="truncate font-medium text-ink">{entry.name}</span>
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
    onColumnWidthChange,
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
    const EmptyIcon = actionIcons.open;
    return (
      <div className="flex flex-col items-center gap-3 rounded-lg border border-hairline bg-surface-card p-8 text-center text-sm text-muted">
        <EmptyIcon className="h-8 w-8 text-muted" aria-hidden />
        <p>This folder is empty. Drop files here or create a subfolder.</p>
      </div>
    );
  }

  return (
    <div className="overflow-x-auto rounded-lg border border-hairline">
      <table className="min-w-full table-fixed divide-y divide-hairline">
        <ListColumnHeader
          listColumns={listColumns}
          entries={entries}
          onColumnToggle={onColumnToggle}
          onColumnWidthChange={onColumnWidthChange}
        />
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
                        className="overflow-hidden px-4 py-3 text-sm text-body"
                        style={{
                          width: column.width,
                          minWidth: column.width,
                          maxWidth: column.width,
                        }}
                        onContextMenu={onContextMenu}
                        onDoubleClick={() => {
                          if (entry.type === "directory") {
                            onOpenDirectory(entry.path);
                          } else {
                            onOpenEntry(entry);
                          }
                        }}
                      >
                        {column.id === "name" ? (
                          renderCell(
                            column,
                            entry,
                            volumeId,
                            renamingPath,
                            onOpenDirectory,
                            onRenameRequest,
                            onRenameCommit,
                            onRenameCancel,
                          )
                        ) : (
                          <span className="block truncate">
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
                          </span>
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
