import { useEffect, useState } from "react";
import { ChevronRight, Trash2 } from "lucide-react";
import { useDroppable } from "@dnd-kit/core";
import { useVolumeFiles } from "@/hooks/useVolumeFiles";
import { treeExpandedPathsForSelection } from "@/lib/explorerPaths";
import { getFileIcon } from "@/lib/icons";
import { cn } from "@/lib/utils";

type TreeNodeProps = {
  volumeId: string;
  path: string;
  name: string;
  depth: number;
  selectedPath: string;
  onSelect: (path: string) => void;
  expanded: Record<string, boolean>;
  onToggle: (path: string) => void;
};

function DroppableTreeNode({
  path,
  children,
  className,
}: {
  path: string;
  children: React.ReactNode;
  className?: string;
}) {
  const { setNodeRef, isOver } = useDroppable({
    id: `tree-folder:${path}`,
    data: { type: "folder", path },
  });
  return (
    <div
      ref={setNodeRef}
      className={cn(className, isOver && "rounded bg-surface-elevated ring-1 ring-primary")}
    >
      {children}
    </div>
  );
}

function TreeNode({
  volumeId,
  path,
  name,
  depth,
  selectedPath,
  onSelect,
  expanded,
  onToggle,
}: TreeNodeProps) {
  const isExpanded = expanded[path] ?? false;
  const listing = useVolumeFiles(volumeId, path);
  const directories =
    listing.data?.entries.filter((entry) => entry.type === "directory") ?? [];
  const normalizedSelected = selectedPath === "" ? "." : selectedPath;
  const isSelected = normalizedSelected === path;

  return (
    <div>
      <DroppableTreeNode path={path}>
        <button
          type="button"
          className={cn(
            "flex w-full items-center gap-1 rounded px-2 py-1.5 text-left text-sm hover:bg-surface-elevated",
            isSelected && "bg-surface-elevated text-primary",
          )}
          style={{ paddingLeft: `${depth * 12 + 8}px` }}
          onClick={() => {
            onSelect(path);
            onToggle(path);
          }}
        >
          <ChevronRight
            className={cn("h-4 w-4 shrink-0 text-muted transition-transform", isExpanded && "rotate-90")}
          />
          {(() => {
            const FolderIcon = getFileIcon({ type: "directory", name });
            return <FolderIcon className="h-4 w-4 shrink-0 text-muted" aria-hidden />;
          })()}
          <span className="truncate">{name}</span>
        </button>
      </DroppableTreeNode>
      {isExpanded
        ? directories.map((dir) => (
            <TreeNode
              key={dir.path}
              volumeId={volumeId}
              path={dir.path}
              name={dir.name}
              depth={depth + 1}
              selectedPath={selectedPath}
              onSelect={onSelect}
              expanded={expanded}
              onToggle={onToggle}
            />
          ))
        : null}
    </div>
  );
}

type FolderTreeProps = {
  volumeId: string;
  selectedPath: string;
  onSelect: (path: string) => void;
  width: number;
  trashSelected?: boolean;
  onSelectTrash?: () => void;
};

export function FolderTree({
  volumeId,
  selectedPath,
  onSelect,
  width,
  trashSelected = false,
  onSelectTrash,
}: FolderTreeProps) {
  const [expanded, setExpanded] = useState<Record<string, boolean>>({ ".": true });

  useEffect(() => {
    const paths = treeExpandedPathsForSelection(selectedPath);
    setExpanded(Object.fromEntries(paths.map((path) => [path, true])));
  }, [selectedPath]);

  return (
    <aside
      className="shrink-0 overflow-auto border-r border-hairline bg-surface-soft"
      style={{ width }}
    >
      <div className="p-2">
        <TreeNode
          volumeId={volumeId}
          path="."
          name="root"
          depth={0}
          selectedPath={selectedPath}
          onSelect={onSelect}
          expanded={expanded}
          onToggle={(path) =>
            setExpanded((prev) => ({ ...prev, [path]: !(prev[path] ?? false) }))
          }
        />
        {onSelectTrash ? (
          <div className="mt-4 border-t border-hairline pt-2">
            <button
              type="button"
              className={cn(
                "flex w-full items-center gap-2 rounded px-2 py-1.5 text-left text-sm hover:bg-surface-elevated",
                trashSelected && "bg-surface-elevated text-primary",
              )}
              onClick={onSelectTrash}
            >
              <Trash2 className="h-4 w-4 shrink-0 text-muted" aria-hidden />
              <span className="text-muted">Trash</span>
            </button>
          </div>
        ) : null}
      </div>
    </aside>
  );
}
