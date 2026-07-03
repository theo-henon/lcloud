import { useEffect, useState } from "react";
import { Header } from "@/components/layout/Header";
import { DiskOverviewCard } from "@/components/monitoring/DiskOverviewCard";
import { UnifiedVolumeList } from "@/components/monitoring/UnifiedVolumeList";
import { VolumeStatsPanel } from "@/components/monitoring/VolumeStatsPanel";
import { Card } from "@/components/ui/card";
import { useMonitoringOverview } from "@/hooks/useMonitoringOverview";
import { useRefreshVolumeStats, useVolumeStats } from "@/hooks/useVolumeStats";
import { navIcons } from "@/lib/icons";

export function MonitoringPage() {
  const overviewQuery = useMonitoringOverview();
  const [selectedVolumeId, setSelectedVolumeId] = useState<string | undefined>();

  const overview = overviewQuery.data;
  const statsQuery = useVolumeStats(selectedVolumeId);
  const refreshStats = useRefreshVolumeStats(selectedVolumeId);

  useEffect(() => {
    if (!selectedVolumeId && overview?.volumes[0]) {
      setSelectedVolumeId(overview.volumes[0].id);
    }
  }, [overview, selectedVolumeId]);

  if (overviewQuery.isLoading) {
    return (
      <>
        <Header
          title="Monitoring"
          description="Track space usage and file breakdown across disks."
        />
        <section className="px-8 py-6">
          <div className="flex flex-col items-center gap-3 text-center">
            {(() => {
              const LoadingIcon = navIcons["/monitoring"];
              return LoadingIcon ? (
                <LoadingIcon className="h-8 w-8 text-muted" aria-hidden />
              ) : null;
            })()}
            <p className="text-sm text-muted">Loading monitoring data...</p>
          </div>
        </section>
      </>
    );
  }

  if (overviewQuery.isError || !overview) {
    return (
      <div className="px-8 py-6">
        <Card className="p-6 text-sm text-accent-rose">Unable to load monitoring dashboard.</Card>
      </div>
    );
  }

  return (
    <>
      <Header
        title="Monitoring"
        description="Track space usage and file breakdown across disks."
      />

      <section className="space-y-8 px-8 py-6">
        <div>
          <h2 className="mb-4 text-sm font-semibold uppercase tracking-wide text-muted">
            Disks
          </h2>
          <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
            {overview.disks.map((disk, index) => (
              <DiskOverviewCard
                key={disk.path || disk.name}
                disk={disk}
                index={index}
              />
            ))}
          </div>
        </div>

        <div>
          <h2 className="mb-4 text-sm font-semibold uppercase tracking-wide text-muted">
            All volumes
          </h2>
          <UnifiedVolumeList
            volumes={overview.volumes}
            disks={overview.disks}
            selectedId={selectedVolumeId}
            onSelect={setSelectedVolumeId}
          />
        </div>

        <VolumeStatsPanel
          stats={statsQuery.data}
          loading={statsQuery.isLoading}
          refreshing={refreshStats.isPending}
          onRefresh={() => {
            void refreshStats.mutateAsync();
          }}
        />
      </section>
    </>
  );
}
