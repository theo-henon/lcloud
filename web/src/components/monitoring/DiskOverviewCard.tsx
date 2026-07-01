import { Card } from "@/components/ui/card";
import type { DiskOverview } from "@/lib/api";
import { formatBytes } from "@/lib/utils";

type DiskOverviewCardProps = {
  disk: DiskOverview;
  index: number;
  maskDiskNames: boolean;
};

export function DiskOverviewCard({ disk, index, maskDiskNames }: DiskOverviewCardProps) {
  const title = maskDiskNames ? `Storage ${index + 1}` : disk.label;
  const usedPercent =
    disk.total_bytes > 0
      ? Math.min(100, (disk.used_by_volumes_bytes / disk.total_bytes) * 100)
      : 0;

  return (
    <Card className="space-y-4 p-5">
      <div>
        <h3 className="text-lg font-semibold text-ink">{title}</h3>
        {!maskDiskNames ? (
          <p className="mt-1 text-sm text-muted">{disk.name}</p>
        ) : null}
      </div>

      <div className="space-y-2">
        <p className="text-2xl font-bold text-primary">
          {formatBytes(disk.free_bytes)} free
        </p>
        <p className="text-sm text-muted">of {formatBytes(disk.total_bytes)} total</p>
      </div>

      <div className="space-y-1">
        <div className="h-2 overflow-hidden rounded-full bg-surface-elevated">
          <div
            className="h-full rounded-full bg-primary"
            style={{ width: `${usedPercent}%` }}
          />
        </div>
        <p className="text-xs text-muted">
          {formatBytes(disk.used_by_volumes_bytes)} used by volumes · {disk.volume_count}{" "}
          volume{disk.volume_count === 1 ? "" : "s"}
        </p>
      </div>
    </Card>
  );
}
