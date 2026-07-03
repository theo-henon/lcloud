import { useDraggable } from "@dnd-kit/core";
import { GripVertical, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

type DraggableWidgetProps = {
  widgetId: string;
  onRemove: () => void;
  children: React.ReactNode;
};

export function DraggableWidget({ widgetId, onRemove, children }: DraggableWidgetProps) {
  const { attributes, listeners, setNodeRef, transform, isDragging } = useDraggable({
    id: widgetId,
    data: { widgetId },
  });

  const style = transform
    ? {
        transform: `translate3d(${transform.x}px, ${transform.y}px, 0)`,
      }
    : undefined;

  return (
    <div
      ref={setNodeRef}
      style={style}
      className={cn("relative h-full", isDragging && "z-20 opacity-70")}
    >
      <div className="absolute right-2 top-2 z-10 flex gap-1">
        <button
          type="button"
          className="inline-flex h-7 w-7 items-center justify-center rounded-md border border-hairline bg-surface-elevated text-muted hover:text-ink"
          aria-label="Drag widget"
          {...listeners}
          {...attributes}
        >
          <GripVertical className="h-4 w-4" aria-hidden />
        </button>
        <Button
          type="button"
          variant="ghost"
          className="h-7 w-7 px-0 text-muted hover:text-accent-rose"
          aria-label="Remove widget"
          onClick={onRemove}
        >
          <X className="h-4 w-4" aria-hidden />
        </Button>
      </div>
      {children}
    </div>
  );
}
