import { useDroppable } from "@dnd-kit/core";
import { cn } from "@/lib/utils";

type GridDropCellProps = {
  x: number;
  y: number;
  visible: boolean;
};

export function GridDropCell({ x, y, visible }: GridDropCellProps) {
  const { setNodeRef, isOver } = useDroppable({
    id: `cell:${x},${y}`,
    data: { x, y },
  });

  if (!visible) {
    return null;
  }

  return (
    <div
      ref={setNodeRef}
      className={cn(
        "pointer-events-auto rounded border border-dashed border-hairline bg-surface-soft/40",
        isOver && "border-primary bg-primary/10",
      )}
      style={{
        gridColumn: `${x + 1}`,
        gridRow: `${y + 1}`,
      }}
      aria-hidden
    />
  );
}
