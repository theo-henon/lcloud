import { COLUMN_LABELS, type ColumnId } from "@/hooks/useExplorerPrefs";
import type { FileEntry } from "@/lib/api";
import { formatBytes } from "@/lib/utils";

export const MIN_COLUMN_WIDTH = 72;
export const MAX_COLUMN_WIDTH = 960;

const HEADER_FONT = "600 12px Inter, sans-serif";
const CELL_FONT = "500 14px Inter, sans-serif";
const CELL_FONT_REGULAR = "400 14px Inter, sans-serif";

let measureContext: CanvasRenderingContext2D | null = null;

function measureTextWidth(text: string, font: string): number {
  if (typeof document === "undefined") {
    return text.length * 8;
  }
  if (!measureContext) {
    measureContext = document.createElement("canvas").getContext("2d");
  }
  if (!measureContext) {
    return text.length * 8;
  }
  measureContext.font = font;
  return measureContext.measureText(text).width;
}

export function clampColumnWidth(width: number): number {
  return Math.min(MAX_COLUMN_WIDTH, Math.max(MIN_COLUMN_WIDTH, Math.round(width)));
}

export function columnCellText(columnId: ColumnId, entry: FileEntry): string {
  switch (columnId) {
    case "name":
      return entry.type === "directory" ? `${entry.name}/` : entry.name;
    case "size":
      return entry.type === "file" ? formatBytes(entry.size_bytes ?? 0) : "—";
    case "modified":
      return entry.modified_at ? new Date(entry.modified_at).toLocaleString() : "—";
    case "type":
      return entry.mime_type ?? (entry.type === "directory" ? "Folder" : "—");
    default:
      return "—";
  }
}

export function autoFitColumnWidth(columnId: ColumnId, entries: FileEntry[]): number {
  const headerPadding = 32;
  const headerWidth =
    measureTextWidth(COLUMN_LABELS[columnId].toUpperCase(), HEADER_FONT) + headerPadding;

  let contentWidth = headerWidth;
  const cellFont = columnId === "name" ? CELL_FONT : CELL_FONT_REGULAR;
  const extraPadding =
    columnId === "name" ? 56 : columnId === "modified" ? 24 : 16;

  for (const entry of entries) {
    const text = columnCellText(columnId, entry);
    contentWidth = Math.max(contentWidth, measureTextWidth(text, cellFont) + extraPadding);
  }

  return clampColumnWidth(contentWidth);
}
