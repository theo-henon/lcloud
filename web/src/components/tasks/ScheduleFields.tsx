import type { TaskScheduleType } from "@/lib/api";

const CRON_PRESETS = [
  { label: "Every Sunday 02:00 UTC", value: "0 2 * * 0" },
  { label: "Daily midnight UTC", value: "0 0 * * *" },
  { label: "Every hour", value: "0 * * * *" },
];

const INTERVAL_PRESETS = [
  { label: "Every hour", value: "1h" },
  { label: "Every 24 hours", value: "24h" },
  { label: "Every 7 days", value: "168h" },
];

type ScheduleFieldsProps = {
  scheduleType: TaskScheduleType;
  schedule: string;
  onScheduleTypeChange: (value: TaskScheduleType) => void;
  onScheduleChange: (value: string) => void;
};

export function ScheduleFields({
  scheduleType,
  schedule,
  onScheduleTypeChange,
  onScheduleChange,
}: ScheduleFieldsProps) {
  const presets = scheduleType === "cron" ? CRON_PRESETS : INTERVAL_PRESETS;

  return (
    <div className="space-y-4">
      <div className="flex gap-4 text-sm">
        <label className="flex items-center gap-2">
          <input
            type="radio"
            checked={scheduleType === "cron"}
            onChange={() => onScheduleTypeChange("cron")}
          />
          Cron (UTC)
        </label>
        <label className="flex items-center gap-2">
          <input
            type="radio"
            checked={scheduleType === "interval"}
            onChange={() => onScheduleTypeChange("interval")}
          />
          Interval
        </label>
      </div>

      <div className="flex flex-wrap gap-2">
        {presets.map((preset) => (
          <button
            key={preset.value}
            type="button"
            className="rounded-md border border-hairline px-3 py-1 text-xs text-body hover:border-primary hover:text-primary"
            onClick={() => onScheduleChange(preset.value)}
          >
            {preset.label}
          </button>
        ))}
      </div>

      <label className="block space-y-1 text-sm">
        <span className="text-muted">
          {scheduleType === "cron" ? "Cron expression" : "Interval (min 1m)"}
        </span>
        <input
          type="text"
          className="w-full rounded-md border border-hairline bg-surface px-3 py-2 font-mono text-ink"
          value={schedule}
          onChange={(e) => onScheduleChange(e.target.value)}
        />
      </label>
    </div>
  );
}
