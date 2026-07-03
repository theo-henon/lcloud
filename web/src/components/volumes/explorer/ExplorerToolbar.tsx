import { useRef } from "react";
import { Button } from "@/components/ui/button";
import { DropdownMenu, DropdownMenuItem } from "@/components/ui/dropdown-menu";
import type { ExplorerPrefs, ViewMode } from "@/hooks/useExplorerPrefs";
import { ActionIcon, actionIcons } from "@/lib/icons";

type ExplorerToolbarProps = {
  prefs: ExplorerPrefs;
  onViewModeChange: (mode: ViewMode) => void;
  onNewFolder: () => void;
  onUpload: (files: FileList) => void;
};

export function ExplorerToolbar({
  prefs,
  onViewModeChange,
  onNewFolder,
  onUpload,
}: ExplorerToolbarProps) {
  const inputRef = useRef<HTMLInputElement>(null);
  const ListIcon = actionIcons.listView;
  const GridIcon = actionIcons.gridView;

  return (
    <div className="flex flex-wrap items-center gap-2 border-b border-hairline px-4 py-3">
      <DropdownMenu
        trigger={
          <Button variant="outline">
            <ActionIcon action="new" className="mr-2" />
            New
          </Button>
        }
      >
        <DropdownMenuItem onSelect={onNewFolder}>
          <ActionIcon action="newFolder" className="mr-2" />
          New folder
        </DropdownMenuItem>
      </DropdownMenu>

      <Button variant="outline" onClick={() => inputRef.current?.click()}>
        <ActionIcon action="upload" className="mr-2" />
        Upload
      </Button>
      <input
        ref={inputRef}
        type="file"
        multiple
        className="hidden"
        onChange={(event) => {
          if (event.target.files?.length) {
            onUpload(event.target.files);
            event.target.value = "";
          }
        }}
      />

      <div className="ml-auto flex items-center gap-2">
        <Button
          variant={prefs.viewMode === "list" ? "default" : "outline"}
          onClick={() => onViewModeChange("list")}
          aria-label="List view"
        >
          <ListIcon className="h-4 w-4" aria-hidden />
        </Button>
        <Button
          variant={prefs.viewMode === "grid" ? "default" : "outline"}
          onClick={() => onViewModeChange("grid")}
          aria-label="Grid view"
        >
          <GridIcon className="h-4 w-4" aria-hidden />
        </Button>
      </div>
    </div>
  );
}
