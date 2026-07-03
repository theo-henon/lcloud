import { Button } from "@/components/ui/button";
import type { CatalogEntry } from "@/lib/dashboard/types";

type WidgetPickerProps = {
  catalog: CatalogEntry[];
  existingTypes: Set<string>;
  onSelect: (type: CatalogEntry["type"]) => void;
  onClose: () => void;
};

export function WidgetPicker({ catalog, existingTypes, onSelect, onClose }: WidgetPickerProps) {
  const available = catalog.filter((entry) => !existingTypes.has(entry.type));

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 p-6">
      <div className="w-full max-w-lg rounded-lg border border-hairline bg-surface-card shadow-xl">
        <div className="flex items-center justify-between border-b border-hairline px-4 py-3">
          <h2 className="text-lg font-semibold text-ink">Add widget</h2>
          <Button variant="ghost" className="h-9 px-3" onClick={onClose}>
            Close
          </Button>
        </div>
        <div className="max-h-[60vh] overflow-auto p-4">
          {available.length === 0 ? (
            <p className="text-sm text-muted">All available widgets are already on your dashboard.</p>
          ) : (
            <ul className="space-y-2">
              {available.map((entry) => (
                  <li key={entry.type}>
                    <button
                      type="button"
                      className="w-full rounded-md border border-hairline bg-surface-soft px-4 py-3 text-left transition-colors hover:border-hairline-strong hover:bg-surface-elevated"
                      onClick={() => onSelect(entry.type)}
                    >
                      <p className="font-medium text-ink">{entry.title}</p>
                      <p className="mt-1 text-sm text-muted">{entry.description}</p>
                      <p className="mt-2 text-xs text-muted">
                        Size: {entry.default_w}×{entry.default_h}
                      </p>
                    </button>
                  </li>
                ))}
            </ul>
          )}
        </div>
      </div>
    </div>
  );
}
