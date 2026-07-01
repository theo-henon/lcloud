import type { CategoryBreakdown } from "@/lib/api";
import { formatBytes } from "@/lib/utils";

const categoryLabels: Record<string, string> = {
  images: "Images",
  videos: "Videos",
  audio: "Audio",
  documents: "Documents",
  archives: "Archives",
  other: "Other",
};

type MimeBreakdownBarsProps = {
  items: CategoryBreakdown[];
};

export function MimeBreakdownBars({ items }: MimeBreakdownBarsProps) {
  if (items.length === 0) {
    return <p className="text-sm text-muted">No files indexed yet.</p>;
  }

  return (
    <div className="space-y-3">
      {items.map((item) => (
        <div key={item.category} className="space-y-1">
          <div className="flex items-center justify-between text-sm">
            <span className="text-body-strong">
              {categoryLabels[item.category] ?? item.category}
            </span>
            <span className="text-muted">
              {formatBytes(item.bytes)} · {Math.round(item.proportion * 100)}%
            </span>
          </div>
          <div className="h-2 overflow-hidden rounded-full bg-surface-elevated">
            <div
              className="h-full rounded-full bg-primary"
              style={{ width: `${Math.max(item.proportion * 100, 1)}%` }}
            />
          </div>
        </div>
      ))}
    </div>
  );
}
