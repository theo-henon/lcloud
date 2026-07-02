import { useCallback } from "react";
import { Columns3 } from "lucide-react";
import { DropdownMenu } from "@/components/ui/dropdown-menu";
import { ColumnPickerPanel } from "@/components/volumes/explorer/ListColumnPicker";
import {
  COLUMN_LABELS,
  type ColumnId,
  type ListColumnPref,
  visibleListColumns,
} from "@/hooks/useExplorerPrefs";
import type { FileEntry } from "@/lib/api";
import { autoFitColumnWidth, clampColumnWidth } from "@/lib/listColumnSizing";
import { cn } from "@/lib/utils";

type ListColumnHeaderProps = {
  listColumns: ListColumnPref[];
  entries: FileEntry[];
  onColumnToggle: (id: ColumnId, visible: boolean) => void;
  onColumnWidthChange: (id: ColumnId, width: number) => void;
};

export function ListColumnHeader({
  listColumns,
  entries,
  onColumnToggle,
  onColumnWidthChange,
}: ListColumnHeaderProps) {
  const visibleColumns = visibleListColumns(listColumns);

  const startResize = useCallback(
    (columnId: ColumnId, startX: number, startWidth: number) => {
      const onMove = (event: MouseEvent) => {
        const nextWidth = clampColumnWidth(startWidth + (event.clientX - startX));
        onColumnWidthChange(columnId, nextWidth);
      };
      const onUp = () => {
        document.removeEventListener("mousemove", onMove);
        document.removeEventListener("mouseup", onUp);
        document.body.style.cursor = "";
        document.body.style.userSelect = "";
      };

      document.body.style.cursor = "col-resize";
      document.body.style.userSelect = "none";
      document.addEventListener("mousemove", onMove);
      document.addEventListener("mouseup", onUp);
    },
    [onColumnWidthChange],
  );

  return (
    <thead className="bg-surface-soft">
      <tr>
        {visibleColumns.map((column) => (
          <th
            key={column.id}
            className="relative px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-muted"
            style={{ width: column.width, minWidth: column.width, maxWidth: column.width }}
            onDoubleClick={() =>
              onColumnWidthChange(column.id, autoFitColumnWidth(column.id, entries))
            }
          >
            <span className="block truncate">{COLUMN_LABELS[column.id]}</span>
            <span
              data-resize-handle
              role="separator"
              aria-orientation="vertical"
              aria-label={`Resize ${COLUMN_LABELS[column.id]} column`}
              className={cn(
                "absolute right-0 top-0 z-10 h-full w-2 translate-x-1/2 cursor-col-resize",
                "hover:bg-primary/40 active:bg-primary/60",
              )}
              onMouseDown={(event) => {
                event.preventDefault();
                event.stopPropagation();
                startResize(column.id, event.clientX, column.width);
              }}
              onDoubleClick={(event) => {
                event.preventDefault();
                event.stopPropagation();
                onColumnWidthChange(column.id, autoFitColumnWidth(column.id, entries));
              }}
            />
          </th>
        ))}
        <th className="w-10 px-2 py-3" style={{ width: 40, minWidth: 40, maxWidth: 40 }}>
          <DropdownMenu
            align="right"
            trigger={
              <button
                type="button"
                className={cn(
                  "inline-flex h-7 w-7 items-center justify-center rounded-md text-muted",
                  "hover:bg-surface-card hover:text-ink",
                )}
                aria-label="Choose columns"
              >
                <Columns3 className="h-4 w-4" />
              </button>
            }
          >
            <ColumnPickerPanel listColumns={listColumns} onColumnToggle={onColumnToggle} />
          </DropdownMenu>
        </th>
      </tr>
    </thead>
  );
}
