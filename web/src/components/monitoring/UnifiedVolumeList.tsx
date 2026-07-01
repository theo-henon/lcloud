import type { DiskOverview, VolumeSummary } from "@/lib/api";
import { formatBytes } from "@/lib/utils";

type UnifiedVolumeListProps = {
  volumes: VolumeSummary[];
  disks: DiskOverview[];
  selectedId?: string;
  onSelect: (volumeId: string) => void;
};

function diskLabel(diskPath: string, disks: DiskOverview[]): string {
  const disk = disks.find((item) => item.path === diskPath);
  return disk?.label ?? diskPath;
}

export function UnifiedVolumeList({
  volumes,
  disks,
  selectedId,
  onSelect,
}: UnifiedVolumeListProps) {
  const showDiskColumn = volumes.some((volume) => volume.disk_path !== "");

  if (volumes.length === 0) {
    return <p className="text-sm text-muted">No volumes yet.</p>;
  }

  return (
    <div className="overflow-hidden rounded-lg border border-hairline">
      <table className="min-w-full text-left text-sm">
        <thead className="bg-surface-soft text-muted">
          <tr>
            <th className="px-4 py-3 font-medium">Volume</th>
            <th className="px-4 py-3 font-medium">Usage</th>
            <th className="px-4 py-3 font-medium">Files</th>
            <th className="px-4 py-3 font-medium">Top type</th>
            {showDiskColumn ? <th className="px-4 py-3 font-medium">Disk</th> : null}
          </tr>
        </thead>
        <tbody>
          {volumes.map((volume) => {
            const usagePercent =
              volume.quota_bytes > 0
                ? Math.min(100, (volume.used_bytes / volume.quota_bytes) * 100)
                : null;
            const isSelected = volume.id === selectedId;

            return (
              <tr
                key={volume.id}
                className={`cursor-pointer border-t border-hairline ${
                  isSelected ? "bg-surface-elevated" : "hover:bg-surface-soft"
                }`}
                onClick={() => onSelect(volume.id)}
              >
                <td className="px-4 py-3 font-medium text-ink">{volume.name}</td>
                <td className="px-4 py-3 text-body">
                  <div className="space-y-1">
                    <span>
                      {volume.quota_bytes > 0
                        ? `${formatBytes(volume.used_bytes)} / ${formatBytes(volume.quota_bytes)}`
                        : `${formatBytes(volume.used_bytes)} / unlimited`}
                    </span>
                    {usagePercent != null ? (
                      <div className="h-1.5 w-32 overflow-hidden rounded-full bg-surface-elevated">
                        <div
                          className="h-full rounded-full bg-primary"
                          style={{ width: `${usagePercent}%` }}
                        />
                      </div>
                    ) : null}
                  </div>
                </td>
                <td className="px-4 py-3 text-body">{volume.file_count}</td>
                <td className="px-4 py-3 capitalize text-body">
                  {volume.top_category ?? "—"}
                </td>
                {showDiskColumn ? (
                  <td className="px-4 py-3 text-muted">
                    {diskLabel(volume.disk_path, disks)}
                  </td>
                ) : null}
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}
