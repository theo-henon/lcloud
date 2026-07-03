import { HardDrive } from "lucide-react";
import { Link } from "react-router-dom";
import { WidgetEmpty, WidgetError, WidgetLoading, WidgetShell } from "@/components/dashboard/WidgetShell";
import { useMonitoringOverview } from "@/hooks/useMonitoringOverview";
import type { WidgetProps } from "@/lib/dashboard/types";
import { formatBytes } from "@/lib/utils";

export function StorageSummaryWidget(_props: WidgetProps) {
  const { data, isLoading, isError } = useMonitoringOverview();

  if (isLoading) {
    return (
      <WidgetShell title="Storage" deepLink="/monitoring">
        <WidgetLoading />
      </WidgetShell>
    );
  }

  if (isError || !data) {
    return (
      <WidgetShell title="Storage" deepLink="/monitoring">
        <WidgetError />
      </WidgetShell>
    );
  }

  const disks = data.disks.slice(0, 2);
  const extraDisks = data.disks.length - disks.length;
  const totalFree = data.disks.reduce((sum, disk) => sum + disk.free_bytes, 0);

  return (
    <WidgetShell title="Storage" deepLink="/monitoring">
      <div className="space-y-3">
        <div>
          <p className="text-2xl font-bold text-primary">{formatBytes(totalFree)}</p>
          <p className="text-xs text-muted">
            free across {data.disks.length} disk{data.disks.length === 1 ? "" : "s"} ·{" "}
            {data.volumes.length} volume{data.volumes.length === 1 ? "" : "s"}
          </p>
        </div>
        {disks.length === 0 ? (
          <WidgetEmpty message="No disks configured yet." />
        ) : (
          <ul className="space-y-2">
            {disks.map((disk) => {
              const usedPercent =
                disk.total_bytes > 0
                  ? Math.min(100, (disk.used_by_volumes_bytes / disk.total_bytes) * 100)
                  : 0;
              return (
                <li key={disk.path || disk.name} className="space-y-1">
                  <div className="flex items-center gap-2 text-sm text-body">
                    <HardDrive className="h-4 w-4 shrink-0 text-muted" aria-hidden />
                    <span className="truncate">{disk.label}</span>
                  </div>
                  <div className="h-1.5 overflow-hidden rounded-full bg-surface-elevated">
                    <div
                      className="h-full rounded-full bg-primary"
                      style={{ width: `${usedPercent}%` }}
                    />
                  </div>
                </li>
              );
            })}
          </ul>
        )}
        {extraDisks > 0 ? (
          <Link to="/monitoring" className="text-xs text-primary hover:text-primary-active">
            +{extraDisks} more disk{extraDisks === 1 ? "" : "s"}
          </Link>
        ) : null}
      </div>
    </WidgetShell>
  );
}
