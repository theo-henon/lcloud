import { Header } from "@/components/layout/Header";
import { PluginActivityFeed } from "@/components/plugins/PluginActivityFeed";
import { PluginList } from "@/components/plugins/PluginList";
import { Card } from "@/components/ui/card";
import { usePluginLogs, usePlugins } from "@/hooks/usePlugins";
import { useAuthStore } from "@/store/auth";

export function PluginsPage() {
  const user = useAuthStore((state) => state.user);
  const isAdmin = user?.role === "admin";
  const pluginsQuery = usePlugins();
  const logsQuery = usePluginLogs({ limit: 50 });

  if (pluginsQuery.isLoading) {
    return (
      <>
        <Header
          title="Plugins"
          description="Extend lcloud with external plugin binaries."
        />
        <section className="px-8 py-6">
          <p className="text-sm text-muted">Loading plugins...</p>
        </section>
      </>
    );
  }

  if (pluginsQuery.isError || !pluginsQuery.data) {
    return (
      <div className="px-8 py-6">
        <Card className="p-6 text-sm text-accent-rose">Unable to load plugin manager.</Card>
      </div>
    );
  }

  return (
    <>
      <Header
        title="Plugins"
        description="Extend lcloud with external plugin binaries."
      />

      <section className="space-y-8 px-8 py-6">
        <div>
          <h2 className="mb-4 text-sm font-semibold uppercase tracking-wide text-muted">
            Installed plugins
          </h2>
          <PluginList plugins={pluginsQuery.data.plugins} isAdmin={isAdmin} />
        </div>

        <div>
          <h2 className="mb-4 text-sm font-semibold uppercase tracking-wide text-muted">
            Activity feed
          </h2>
          {logsQuery.isError ? (
            <Card className="p-6 text-sm text-accent-rose">Unable to load plugin activity.</Card>
          ) : (
            <PluginActivityFeed entries={logsQuery.data?.entries ?? []} />
          )}
        </div>
      </section>
    </>
  );
}
