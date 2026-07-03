import { HardDrive } from "lucide-react";
import type { DiskOverview, VolumeSummary } from "@/lib/api";
import { navIcons } from "@/lib/icons";
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
  const showOwnerColumn = volumes.some((volume) => volume.owner_email);

  if (volumes.length === 0) {
    const EmptyIcon = navIcons["/volumes"];
    return (
      <div className="flex flex-col items-center gap-3 py-6 text-center text-sm text-muted">
        {EmptyIcon ? <EmptyIcon className="h-8 w-8 text-muted" aria-hidden /> : null}
        <p>No volumes yet.</p>
      </div>
    );
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
            {showOwnerColumn ? <th className="px-4 py-3 font-medium">Owner</th> : null}
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
                <td className="px-4 py-3 font-medium text-ink">
                  <span className="flex items-center gap-2">
                    <HardDrive className="h-4 w-4 shrink-0 text-muted" aria-hidden />
                    {volume.name}
                  </span>
                </td>
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
                {showOwnerColumn ? (
                  <td className="px-4 py-3 text-body">{volume.owner_email ?? "—"}</td>
                ) : null}
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
