import { WidgetEmpty, WidgetError, WidgetLoading, WidgetShell } from "@/components/dashboard/WidgetShell";
import { useMonitoringOverview } from "@/hooks/useMonitoringOverview";
import type { WidgetProps } from "@/lib/dashboard/types";
import { formatBytes } from "@/lib/utils";

const RISK_THRESHOLD = 0.85;

export function VolumesAtRiskWidget(_props: WidgetProps) {
  const { data, isLoading, isError } = useMonitoringOverview();

  if (isLoading) {
    return (
      <WidgetShell title="Volumes at risk" deepLink="/monitoring">
        <WidgetLoading />
      </WidgetShell>
    );
  }

  if (isError || !data) {
    return (
      <WidgetShell title="Volumes at risk" deepLink="/monitoring">
        <WidgetError />
      </WidgetShell>
    );
  }

  const atRisk = data.volumes
    .filter(
      (volume) =>
        volume.quota_bytes > 0 &&
        volume.used_bytes / volume.quota_bytes >= RISK_THRESHOLD,
    )
    .slice(0, 5);

  if (atRisk.length === 0) {
    return (
      <WidgetShell title="Volumes at risk" deepLink="/monitoring">
        <WidgetEmpty message="All volumes healthy." />
      </WidgetShell>
    );
  }

  return (
    <WidgetShell title="Volumes at risk" deepLink="/monitoring">
      <ul className="space-y-3">
        {atRisk.map((volume) => {
          const percent = Math.min(
            100,
            volume.quota_bytes > 0 ? (volume.used_bytes / volume.quota_bytes) * 100 : 0,
          );
          return (
            <li key={volume.id} className="space-y-1">
              <div className="flex items-center justify-between gap-2 text-sm">
                <span className="truncate font-medium text-ink">{volume.name}</span>
                <span className="shrink-0 text-muted">{percent.toFixed(0)}%</span>
              </div>
              <div className="h-1.5 overflow-hidden rounded-full bg-surface-elevated">
                <div
                  className="h-full rounded-full bg-accent-rose"
                  style={{ width: `${percent}%` }}
                />
              </div>
              <p className="text-xs text-muted">
                {formatBytes(volume.used_bytes)} / {formatBytes(volume.quota_bytes)}
              </p>
            </li>
          );
        })}
      </ul>
    </WidgetShell>
  );
}
