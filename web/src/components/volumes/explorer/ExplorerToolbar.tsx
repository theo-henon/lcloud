import { useRef } from "react";
import { FolderPlus, Grid3X3, List, Upload } from "lucide-react";
import { Button } from "@/components/ui/button";
import { DropdownMenu, DropdownMenuItem } from "@/components/ui/dropdown-menu";
import type { ExplorerPrefs, ViewMode } from "@/hooks/useExplorerPrefs";

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

  return (
    <div className="flex flex-wrap items-center gap-2 border-b border-hairline px-4 py-3">
      <DropdownMenu
        trigger={
          <Button variant="outline">
            <FolderPlus className="mr-2 h-4 w-4" />
            New
          </Button>
        }
      >
        <DropdownMenuItem onSelect={onNewFolder}>New folder</DropdownMenuItem>
      </DropdownMenu>

      <Button variant="outline" onClick={() => inputRef.current?.click()}>
        <Upload className="mr-2 h-4 w-4" />
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
          <List className="h-4 w-4" />
        </Button>
        <Button
          variant={prefs.viewMode === "grid" ? "default" : "outline"}
          onClick={() => onViewModeChange("grid")}
          aria-label="Grid view"
        >
          <Grid3X3 className="h-4 w-4" />
        </Button>
      </div>
    </div>
  );
}
