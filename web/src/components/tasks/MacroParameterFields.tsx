export const MACRO_GROUPS = [
  {
    label: "Cleanup",
    macros: [
      { id: "delete_old_files", label: "Delete old files" },
      { id: "delete_large_files", label: "Delete large files" },
      { id: "clear_cache", label: "Clear cache" },
    ],
  },
  {
    label: "Organization",
    macros: [
      { id: "move_files", label: "Move matching files" },
      { id: "sort_by_type", label: "Sort by file type" },
      { id: "sort_by_date", label: "Sort by date" },
    ],
  },
  {
    label: "Maintenance",
    macros: [
      { id: "rebuild_index", label: "Rebuild search index" },
      { id: "compute_stats", label: "Recompute statistics" },
    ],
  },
  {
    label: "Alerts",
    macros: [{ id: "alert_usage", label: "Usage alert" }],
  },
] as const;

export const DESTRUCTIVE_MACROS = new Set([
  "delete_old_files",
  "delete_large_files",
  "clear_cache",
  "move_files",
  "sort_by_type",
  "sort_by_date",
]);

export const GLOBAL_MACROS = new Set(["rebuild_index", "compute_stats", "alert_usage"]);

export function macroLabel(macro: string): string {
  for (const group of MACRO_GROUPS) {
    const found = group.macros.find((item) => item.id === macro);
    if (found) return found.label;
  }
  return macro;
}

type MacroParameterFieldsProps = {
  macro: string;
  parameters: Record<string, unknown>;
  onChange: (parameters: Record<string, unknown>) => void;
};

export function MacroParameterFields({
  macro,
  parameters,
  onChange,
}: MacroParameterFieldsProps) {
  const setParam = (key: string, value: unknown) => {
    onChange({ ...parameters, [key]: value });
  };

  const dryRun = Boolean(parameters.dry_run);
  const showDryRun = DESTRUCTIVE_MACROS.has(macro);

  return (
    <div className="space-y-4">
      {macro === "delete_old_files" ? (
        <>
          <label className="block space-y-1 text-sm">
            <span className="text-muted">Days</span>
            <input
              type="number"
              min={1}
              className="w-full rounded-md border border-hairline bg-surface px-3 py-2 text-ink"
              value={Number(parameters.days ?? 90)}
              onChange={(e) => setParam("days", Number(e.target.value))}
            />
          </label>
          <label className="block space-y-1 text-sm">
            <span className="text-muted">Extensions (optional, comma-separated)</span>
            <input
              type="text"
              placeholder=".jpg, .png"
              className="w-full rounded-md border border-hairline bg-surface px-3 py-2 text-ink"
              value={Array.isArray(parameters.extensions) ? (parameters.extensions as string[]).join(", ") : ""}
              onChange={(e) =>
                setParam(
                  "extensions",
                  e.target.value
                    .split(",")
                    .map((s) => s.trim())
                    .filter(Boolean),
                )
              }
            />
          </label>
        </>
      ) : null}

      {macro === "delete_large_files" ? (
        <>
          <label className="block space-y-1 text-sm">
            <span className="text-muted">Minimum size (MB)</span>
            <input
              type="number"
              min={1}
              className="w-full rounded-md border border-hairline bg-surface px-3 py-2 text-ink"
              value={Number(parameters.min_size_mb ?? 100)}
              onChange={(e) => setParam("min_size_mb", Number(e.target.value))}
            />
          </label>
        </>
      ) : null}

      {macro === "move_files" ? (
        <>
          <label className="block space-y-1 text-sm">
            <span className="text-muted">Glob pattern</span>
            <input
              type="text"
              placeholder="*.jpg"
              className="w-full rounded-md border border-hairline bg-surface px-3 py-2 font-mono text-ink"
              value={String(parameters.pattern ?? "")}
              onChange={(e) => setParam("pattern", e.target.value)}
            />
          </label>
          <label className="block space-y-1 text-sm">
            <span className="text-muted">Target subfolder</span>
            <input
              type="text"
              placeholder="archive"
              className="w-full rounded-md border border-hairline bg-surface px-3 py-2 text-ink"
              value={String(parameters.target_subfolder ?? "")}
              onChange={(e) => setParam("target_subfolder", e.target.value)}
            />
          </label>
        </>
      ) : null}

      {macro === "alert_usage" ? (
        <label className="block space-y-1 text-sm">
          <span className="text-muted">Threshold (%)</span>
          <input
            type="number"
            min={1}
            max={100}
            className="w-full rounded-md border border-hairline bg-surface px-3 py-2 text-ink"
            value={Number(parameters.threshold_percent ?? 80)}
            onChange={(e) => setParam("threshold_percent", Number(e.target.value))}
          />
        </label>
      ) : null}

      {showDryRun ? (
        <label className="flex items-start gap-2 text-sm text-body">
          <input
            type="checkbox"
            checked={dryRun}
            onChange={(e) => setParam("dry_run", e.target.checked)}
            className="mt-1"
          />
          <span>
            Simulation only (dry-run)
            <span className="mt-1 block text-xs text-muted">
              Aucun fichier ne sera modifié — prévisualisation uniquement.
            </span>
          </span>
        </label>
      ) : null}
    </div>
  );
}
