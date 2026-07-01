import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { MimeBreakdownBars } from "@/components/monitoring/MimeBreakdownBars";
import type { VolumeStats } from "@/lib/api";
import { formatBytes } from "@/lib/utils";

type VolumeStatsPanelProps = {
  stats: VolumeStats | undefined;
  loading: boolean;
  refreshing: boolean;
  onRefresh: () => void;
};

export function VolumeStatsPanel({
  stats,
  loading,
  refreshing,
  onRefresh,
}: VolumeStatsPanelProps) {
  if (loading) {
    return <Card className="p-6 text-sm text-muted">Loading volume stats...</Card>;
  }

  if (!stats) {
    return (
      <Card className="p-6 text-sm text-muted">
        Select a volume to see detailed statistics.
      </Card>
    );
  }

  const usageLabel =
    stats.quota_bytes > 0 && stats.usage_percent != null
      ? `${Math.round(stats.usage_percent * 100)}%`
      : null;

  return (
    <Card className="space-y-6 p-6">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <h3 className="text-xl font-semibold text-ink">{stats.name}</h3>
          <p className="mt-1 text-sm text-muted">
            {stats.file_count} file{stats.file_count === 1 ? "" : "s"} · updated{" "}
            {new Date(stats.computed_at).toLocaleString()}
          </p>
        </div>
        <Button variant="outline" disabled={refreshing} onClick={onRefresh}>
          {refreshing ? "Refreshing..." : "Refresh stats"}
        </Button>
      </div>

      <div className="grid gap-4 sm:grid-cols-3">
        <div>
          <p className="text-sm text-muted">Used</p>
          <p className="text-3xl font-bold text-primary">{formatBytes(stats.used_bytes)}</p>
        </div>
        <div>
          <p className="text-sm text-muted">Quota</p>
          <p className="text-3xl font-bold text-primary">
            {stats.quota_bytes > 0 ? formatBytes(stats.quota_bytes) : "Unlimited"}
          </p>
        </div>
        <div>
          <p className="text-sm text-muted">Usage</p>
          <p className="text-3xl font-bold text-primary">{usageLabel ?? "—"}</p>
        </div>
      </div>

      <div>
        <h4 className="mb-3 text-sm font-semibold text-body-strong">By category</h4>
        <MimeBreakdownBars items={stats.by_category} />
      </div>

      {stats.by_mime.length > 0 ? (
        <div>
          <h4 className="mb-3 text-sm font-semibold text-body-strong">Top MIME types</h4>
          <ul className="space-y-2 text-sm text-body">
            {stats.by_mime.slice(0, 5).map((item) => (
              <li key={item.mime_type} className="flex justify-between gap-4">
                <span>{item.mime_type}</span>
                <span className="text-muted">
                  {formatBytes(item.bytes)} · {Math.round(item.proportion * 100)}%
                </span>
              </li>
            ))}
          </ul>
        </div>
      ) : null}
    </Card>
  );
}
