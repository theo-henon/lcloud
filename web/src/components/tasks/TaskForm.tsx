import { useState } from "react";
import type { TaskRecord, Volume } from "@/lib/api";
import {
  DESTRUCTIVE_MACROS,
  GLOBAL_MACROS,
  MACRO_GROUPS,
  MacroParameterFields,
} from "@/components/tasks/MacroParameterFields";
import { ScheduleFields } from "@/components/tasks/ScheduleFields";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";

type TaskFormProps = {
  volumes: Volume[];
  isAdmin: boolean;
  initial?: TaskRecord;
  onSubmit: (payload: {
    name: string;
    macro: string;
    scope: "volume" | "global";
    volume_id?: string;
    parameters: Record<string, unknown>;
    schedule_type: "cron" | "interval";
    schedule: string;
    enabled: boolean;
  }) => void;
  onCancel: () => void;
  isPending?: boolean;
};

export function TaskForm({
  volumes,
  isAdmin,
  initial,
  onSubmit,
  onCancel,
  isPending,
}: TaskFormProps) {
  const [name, setName] = useState(initial?.name ?? "");
  const [macro, setMacro] = useState(initial?.macro ?? "delete_old_files");
  const [scope, setScope] = useState<"volume" | "global">(initial?.scope ?? "volume");
  const [volumeId, setVolumeId] = useState(initial?.volume_id ?? volumes[0]?.id ?? "");
  const [parameters, setParameters] = useState<Record<string, unknown>>(
    initial?.parameters ?? { days: 90 },
  );
  const [scheduleType, setScheduleType] = useState<"cron" | "interval">(
    initial?.schedule_type ?? "cron",
  );
  const [schedule, setSchedule] = useState(initial?.schedule ?? "0 2 * * 0");
  const [enabled, setEnabled] = useState(initial?.enabled ?? true);

  const handleMacroChange = (nextMacro: string) => {
    setMacro(nextMacro);
    if (GLOBAL_MACROS.has(nextMacro) && isAdmin) {
      setScope("global");
    } else {
      setScope("volume");
    }
    if (nextMacro === "delete_old_files") setParameters({ days: 90 });
    else if (nextMacro === "purge_trash") setParameters({ days: 30 });
    else if (nextMacro === "delete_large_files") setParameters({ min_size_mb: 100 });
    else if (nextMacro === "move_files") setParameters({ pattern: "*.jpg", target_subfolder: "archive" });
    else if (nextMacro === "alert_usage") setParameters({ threshold_percent: 80 });
    else setParameters({});
  };

  return (
    <Card className="p-6">
      <h3 className="mb-4 text-base font-semibold text-ink">
        {initial ? "Edit task" : "Create task"}
      </h3>
      <form
        className="space-y-5"
        onSubmit={(e) => {
          e.preventDefault();
          onSubmit({
            name,
            macro,
            scope,
            volume_id: scope === "volume" ? volumeId : undefined,
            parameters,
            schedule_type: scheduleType,
            schedule,
            enabled,
          });
        }}
      >
        <label className="block space-y-1 text-sm">
          <span className="text-muted">Name</span>
          <input
            required
            className="w-full rounded-md border border-hairline bg-surface px-3 py-2 text-ink"
            value={name}
            onChange={(e) => setName(e.target.value)}
          />
        </label>

        <label className="block space-y-1 text-sm">
          <span className="text-muted">Macro</span>
          <select
            className="w-full rounded-md border border-hairline bg-surface px-3 py-2 text-ink"
            value={macro}
            onChange={(e) => handleMacroChange(e.target.value)}
          >
            {MACRO_GROUPS.map((group) => (
              <optgroup key={group.label} label={group.label}>
                {group.macros.map((item) => (
                  <option key={item.id} value={item.id}>
                    {item.label}
                  </option>
                ))}
              </optgroup>
            ))}
          </select>
        </label>

        {scope === "volume" ? (
          <label className="block space-y-1 text-sm">
            <span className="text-muted">Volume</span>
            <select
              required
              className="w-full rounded-md border border-hairline bg-surface px-3 py-2 text-ink"
              value={volumeId}
              onChange={(e) => setVolumeId(e.target.value)}
            >
              {volumes.map((vol) => (
                <option key={vol.id} value={vol.id}>
                  {vol.name}
                </option>
              ))}
            </select>
          </label>
        ) : (
          <p className="text-sm text-muted">Global task — runs on all volumes (admin only).</p>
        )}

        <MacroParameterFields macro={macro} parameters={parameters} onChange={setParameters} />
        <ScheduleFields
          scheduleType={scheduleType}
          schedule={schedule}
          onScheduleTypeChange={setScheduleType}
          onScheduleChange={setSchedule}
        />

        <label className="flex items-center gap-2 text-sm text-body">
          <input
            type="checkbox"
            checked={enabled}
            onChange={(e) => setEnabled(e.target.checked)}
          />
          Enabled
        </label>

        {DESTRUCTIVE_MACROS.has(macro) && !parameters.dry_run ? (
          <p className="text-xs text-muted">
            Cette tâche modifiera des fichiers lors de l&apos;exécution planifiée.
          </p>
        ) : null}

        <div className="flex gap-2">
          <Button type="submit" disabled={isPending}>
            {initial ? "Save" : "Create"}
          </Button>
          <Button type="button" variant="outline" onClick={onCancel}>
            Cancel
          </Button>
        </div>
      </form>
    </Card>
  );
}
