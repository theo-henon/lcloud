import { useCallback, useEffect, useState, type ReactNode } from "react";
import {
  DndContext,
  DragOverlay,
  PointerSensor,
  useDroppable,
  useSensor,
  useSensors,
  type DragEndEvent,
} from "@dnd-kit/core";
import { useQueryClient } from "@tanstack/react-query";
import { BreadcrumbNav } from "@/components/volumes/BreadcrumbNav";
import { ExplorerToolbar } from "@/components/volumes/explorer/ExplorerToolbar";
import { FileGridView } from "@/components/volumes/explorer/FileGridView";
import { FileListView } from "@/components/volumes/explorer/FileListView";
import { FolderTree } from "@/components/volumes/explorer/FolderTree";
import { PreviewPanel } from "@/components/volumes/explorer/PreviewPanel";
import { UploadQueue } from "@/components/volumes/explorer/UploadQueue";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { useExplorerPrefs, visibleColumns, type ColumnId } from "@/hooks/useExplorerPrefs";
import { useUploadQueue } from "@/hooks/useUploadQueue";
import {
  useCreateDirectory,
  useDeleteFile,
  useMoveFile,
  useRenameFile,
  useVolumeFiles,
} from "@/hooks/useVolumeFiles";
import "@/lib/fileOpeners";
import { api } from "@/lib/api";
import type { FileEntry } from "@/lib/api";
import { joinPath, fileName as baseName } from "@/lib/explorerPaths";
import { resolveFileOpener } from "@/lib/fileOpeners/registry";
import { useAuthStore } from "@/store/auth";
import { cn } from "@/lib/utils";

type VolumeExplorerProps = {
  volumeId: string;
  currentPath: string;
  onPathChange: (path: string) => void;
};

function ContentsDropZone({
  currentPath,
  onExternalDrop,
  children,
}: {
  currentPath: string;
  onExternalDrop: (files: FileList, targetPath: string) => void;
  children: ReactNode;
}) {
  const { setNodeRef, isOver } = useDroppable({
    id: `contents:${currentPath}`,
    data: { type: "folder", path: currentPath },
  });
  const [dragOver, setDragOver] = useState(false);

  return (
    <div
      ref={setNodeRef}
      className={cn(
        "min-h-[320px] rounded-lg p-1 transition-colors",
        (dragOver || isOver) && "bg-surface-elevated ring-2 ring-primary ring-offset-2 ring-offset-canvas",
      )}
      onDragOver={(event) => {
        if (event.dataTransfer.types.includes("Files")) {
          event.preventDefault();
          setDragOver(true);
        }
      }}
      onDragLeave={() => setDragOver(false)}
      onDrop={(event) => {
        event.preventDefault();
        setDragOver(false);
        if (event.dataTransfer.files.length > 0) {
          onExternalDrop(event.dataTransfer.files, currentPath);
        }
      }}
    >
      {children}
    </div>
  );
}

export function VolumeExplorer({ volumeId, currentPath, onPathChange }: VolumeExplorerProps) {
  const accessToken = useAuthStore((state) => state.accessToken);
  const queryClient = useQueryClient();
  const { prefs, setPrefs } = useExplorerPrefs();
  const filesQuery = useVolumeFiles(volumeId, currentPath);
  const createDirectory = useCreateDirectory(volumeId);
  const deleteFile = useDeleteFile(volumeId);
  const moveFile = useMoveFile(volumeId);
  const renameFile = useRenameFile(volumeId);
  const invalidateFiles = useCallback(() => {
    void queryClient.invalidateQueries({ queryKey: ["volumes", volumeId, "files"] });
  }, [queryClient, volumeId]);

  const { items, enqueueFiles, dismissItem, cancelQueued } = useUploadQueue(volumeId, invalidateFiles);

  const [showNewFolder, setShowNewFolder] = useState(false);
  const [newFolderName, setNewFolderName] = useState("");
  const [renamingPath, setRenamingPath] = useState<string | null>(null);
  const [preview, setPreview] = useState<{ entry: FileEntry; content: ReactNode } | null>(null);

  const sensors = useSensors(useSensor(PointerSensor, { activationConstraint: { distance: 6 } }));

  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        setPreview(null);
        setRenamingPath(null);
      }
    };
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, []);

  const handleOpenEntry = useCallback(
    async (entry: FileEntry) => {
      if (entry.type === "directory") {
        onPathChange(entry.path);
        return;
      }
      const opener = resolveFileOpener(entry);
      if (!opener) {
        void api.downloadFile(volumeId, entry.path, entry.name);
        return;
      }
      const contentUrl = api.fileContentUrl(volumeId, entry.path, "inline");
      await opener.open(
        { volumeId, entry, contentUrl, accessToken },
        {
          render: (node) => setPreview({ entry, content: node }),
          close: () => setPreview(null),
        },
      );
    },
    [accessToken, onPathChange, volumeId],
  );

  const handleRenameCommit = useCallback(
    async (entry: FileEntry, newName: string) => {
      setRenamingPath(null);
      if (!newName || newName === entry.name) {
        return;
      }
      try {
        await renameFile.mutateAsync({ path: entry.path, newName });
      } catch {
        setRenamingPath(entry.path);
      }
    },
    [renameFile],
  );

  const handleDragEnd = useCallback(
    (event: DragEndEvent) => {
      const fromPath = event.active.data.current?.path as string | undefined;
      const overType = event.over?.data.current?.type as string | undefined;
      const toFolder = event.over?.data.current?.path as string | undefined;
      if (!fromPath || overType !== "folder" || !toFolder) {
        return;
      }
      const destination = joinPath(toFolder === "." ? "." : toFolder, baseName(fromPath));
      if (destination === fromPath) {
        return;
      }
      void moveFile.mutateAsync({ fromPath, toPath: destination });
    },
    [moveFile],
  );

  const entries = filesQuery.data?.entries ?? [];
  const columns = visibleColumns(prefs);

  return (
    <DndContext sensors={sensors} onDragEnd={handleDragEnd}>
      <div className="flex min-h-[560px] overflow-hidden rounded-lg border border-hairline bg-surface-card">
        <FolderTree
          volumeId={volumeId}
          selectedPath={currentPath}
          onSelect={onPathChange}
          width={prefs.treeWidth}
        />
        <div className="flex min-w-0 flex-1 flex-col">
          <ExplorerToolbar
            prefs={prefs}
            onViewModeChange={(viewMode) => setPrefs((prev) => ({ ...prev, viewMode }))}
            onColumnToggle={(id: ColumnId, visible) =>
              setPrefs((prev) => ({
                ...prev,
                listColumns: prev.listColumns.map((column) =>
                  column.id === id ? { ...column, visible } : column,
                ),
              }))
            }
            onNewFolder={() => setShowNewFolder(true)}
            onUpload={(files) => enqueueFiles(files, currentPath)}
          />

          <div className="space-y-4 p-4">
            <BreadcrumbNav
              volumeId={volumeId}
              path={currentPath}
              onNavigate={(path) => onPathChange(path || ".")}
            />

            <UploadQueue items={items} onDismiss={dismissItem} onCancel={cancelQueued} />

            {showNewFolder ? (
              <Card className="flex flex-wrap items-end gap-3 p-4">
                <div className="min-w-[220px] flex-1 space-y-2">
                  <label className="text-sm font-medium text-body-strong" htmlFor="new-folder-name">
                    Folder name
                  </label>
                  <Input
                    id="new-folder-name"
                    value={newFolderName}
                    onChange={(event) => setNewFolderName(event.target.value)}
                    placeholder="2024/vacation"
                  />
                </div>
                <Button
                  disabled={!newFolderName.trim() || createDirectory.isPending}
                  onClick={() => {
                    const target = joinPath(currentPath, newFolderName.trim());
                    void createDirectory.mutateAsync(target).then(() => {
                      setNewFolderName("");
                      setShowNewFolder(false);
                    });
                  }}
                >
                  Create
                </Button>
                <Button variant="ghost" onClick={() => setShowNewFolder(false)}>
                  Cancel
                </Button>
              </Card>
            ) : null}

            <ContentsDropZone
              currentPath={currentPath}
              onExternalDrop={(files, targetPath) => enqueueFiles(files, targetPath)}
            >
              {filesQuery.isLoading ? (
                <p className="text-sm text-muted">Loading files…</p>
              ) : prefs.viewMode === "grid" ? (
                <FileGridView
                  volumeId={volumeId}
                  entries={entries}
                  renamingPath={renamingPath}
                  onOpenDirectory={onPathChange}
                  onOpenEntry={(entry) => void handleOpenEntry(entry)}
                  onDownload={(entry) => void api.downloadFile(volumeId, entry.path, entry.name)}
                  onDelete={(entry) => {
                    if (window.confirm("Delete this file?")) {
                      void deleteFile.mutateAsync(entry.path);
                    }
                  }}
                  onRenameRequest={(entry) => setRenamingPath(entry.path)}
                  onRenameCommit={(entry, newName) => void handleRenameCommit(entry, newName)}
                  onRenameCancel={() => setRenamingPath(null)}
                />
              ) : (
                <FileListView
                  volumeId={volumeId}
                  entries={entries}
                  columns={columns}
                  renamingPath={renamingPath}
                  onOpenDirectory={onPathChange}
                  onOpenEntry={(entry) => void handleOpenEntry(entry)}
                  onDownload={(entry) => void api.downloadFile(volumeId, entry.path, entry.name)}
                  onDelete={(entry) => {
                    if (window.confirm("Delete this file?")) {
                      void deleteFile.mutateAsync(entry.path);
                    }
                  }}
                  onRenameRequest={(entry) => setRenamingPath(entry.path)}
                  onRenameCommit={(entry, newName) => void handleRenameCommit(entry, newName)}
                  onRenameCancel={() => setRenamingPath(null)}
                />
              )}
            </ContentsDropZone>
          </div>
        </div>
      </div>

      <DragOverlay />

      {preview ? (
        <PreviewPanel
          entry={preview.entry}
          volumeId={volumeId}
          content={preview.content}
          onClose={() => setPreview(null)}
        />
      ) : null}
    </DndContext>
  );
}
