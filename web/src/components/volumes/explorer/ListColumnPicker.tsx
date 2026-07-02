import {
  DropdownMenuCheckboxItem,
} from "@/components/ui/dropdown-menu";
import {
  COLUMN_LABELS,
  type ColumnId,
  type ListColumnPref,
} from "@/hooks/useExplorerPrefs";

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
