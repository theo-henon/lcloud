import type { WidgetPlacement, WidgetType } from "@/lib/dashboard/types";

export function gridBottom(widgets: WidgetPlacement[]): number {
  if (widgets.length === 0) {
    return 0;
  }
  return widgets.reduce((max, widget) => Math.max(max, widget.y + widget.h), 0);
}

export function nextRow(widgets: WidgetPlacement[]): number {
  return gridBottom(widgets);
}

export function widgetsOverlap(a: WidgetPlacement, b: WidgetPlacement): boolean {
  return a.x < b.x + b.w && a.x + a.w > b.x && a.y < b.y + b.h && a.y + a.h > b.y;
}

export function hasOverlap(widgets: WidgetPlacement[], candidate: WidgetPlacement): boolean {
  return widgets.some(
    (widget) => widget.id !== candidate.id && widgetsOverlap(widget, candidate),
  );
}

export function moveWidget(
  widgets: WidgetPlacement[],
  widgetId: string,
  x: number,
  y: number,
): WidgetPlacement[] {
  return widgets.map((widget) =>
    widget.id === widgetId ? { ...widget, x, y } : widget,
  );
}

export function removeWidget(widgets: WidgetPlacement[], widgetId: string): WidgetPlacement[] {
  return widgets.filter((widget) => widget.id !== widgetId);
}

export function addWidget(
  widgets: WidgetPlacement[],
  type: WidgetType,
  w: number,
  h: number,
): WidgetPlacement[] {
  const y = nextRow(widgets);
  return [
    ...widgets,
    {
      id: crypto.randomUUID(),
      type,
      x: 0,
      y,
      w,
      h,
    },
  ];
}

export function gridExtent(widgets: WidgetPlacement[]): { cols: number; rows: number } {
  const rows = Math.max(gridBottom(widgets), 4);
  return { cols: 12, rows };
}
