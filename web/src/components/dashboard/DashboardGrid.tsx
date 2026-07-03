import type { WidgetPlacement } from "@/lib/dashboard/types";
import { getWidgetDefinition } from "@/lib/dashboard/registry";
import { DraggableWidget } from "@/components/dashboard/DraggableWidget";
import { GridDropCell } from "@/components/dashboard/GridDropCell";

type DashboardGridProps = {
  widgets: WidgetPlacement[];
  editing?: boolean;
  onRemoveWidget?: (widgetId: string) => void;
  showDropCells?: boolean;
  maxRows?: number;
};

function renderWidget(
  widget: WidgetPlacement,
  editing: boolean,
  onRemoveWidget?: (widgetId: string) => void,
) {
  const definition = getWidgetDefinition(widget.type);
  const Component = definition.component;
  const content = <Component id={widget.id} type={widget.type} />;

  if (editing && onRemoveWidget) {
    return (
      <DraggableWidget widgetId={widget.id} onRemove={() => onRemoveWidget(widget.id)}>
        {content}
      </DraggableWidget>
    );
  }

  return content;
}

export function DashboardGrid({
  widgets,
  editing = false,
  onRemoveWidget,
  showDropCells = false,
  maxRows = 8,
}: DashboardGridProps) {
  const sorted = [...widgets].sort((a, b) => a.y - b.y || a.x - b.x);

  return (
    <div
      className="relative grid grid-cols-1 gap-4 md:grid-cols-12 md:auto-rows-[minmax(88px,auto)]"
      style={
        showDropCells
          ? { gridTemplateRows: `repeat(${maxRows}, minmax(88px, auto))` }
          : undefined
      }
    >
      {showDropCells
        ? Array.from({ length: maxRows }, (_, row) =>
            Array.from({ length: 12 }, (_, col) => (
              <GridDropCell key={`${col}-${row}`} x={col} y={row} visible />
            )),
          )
        : null}

      {sorted.map((widget) => (
        <div
          key={widget.id}
          className="relative z-10 min-h-[88px] md:[grid-column:var(--widget-col)] md:[grid-row:var(--widget-row)]"
          style={{
            ["--widget-col" as string]: `${widget.x + 1} / span ${widget.w}`,
            ["--widget-row" as string]: `${widget.y + 1} / span ${widget.h}`,
          }}
        >
          {renderWidget(widget, editing, onRemoveWidget)}
        </div>
      ))}
    </div>
  );
}
