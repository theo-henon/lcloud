import { Columns3 } from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
} from "@/components/ui/dropdown-menu";
import {
  COLUMN_LABELS,
  type ColumnId,
  type ListColumnPref,
  visibleListColumns,
} from "@/hooks/useExplorerPrefs";
import { cn } from "@/lib/utils";

type ListColumnPickerProps = {
  listColumns: ListColumnPref[];
  onColumnToggle: (id: ColumnId, visible: boolean) => void;
};

function sortedColumns(listColumns: ListColumnPref[]) {
  return [...listColumns].sort((a, b) => a.order - b.order);
}

export function ColumnPickerPanel({
  listColumns,
  onColumnToggle,
}: ListColumnPickerProps) {
  return (
    <>
      {sortedColumns(listColumns).map((column) => (
        <DropdownMenuCheckboxItem
          key={column.id}
          checked={column.visible}
          label={COLUMN_LABELS[column.id]}
          disabled={column.id === "name"}
          onCheckedChange={(checked) => onColumnToggle(column.id, checked)}
        />
      ))}
    </>
  );
}

export function ListColumnHeader({
  listColumns,
  onColumnToggle,
}: ListColumnPickerProps) {
  const visibleColumns = visibleListColumns(listColumns);

  return (
    <thead className="bg-surface-soft">
      <tr>
        {visibleColumns.map((column) => (
          <th
            key={column.id}
            className="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-muted"
            style={{ width: column.width }}
          >
            {COLUMN_LABELS[column.id]}
          </th>
        ))}
        <th className="w-10 px-2 py-3">
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
