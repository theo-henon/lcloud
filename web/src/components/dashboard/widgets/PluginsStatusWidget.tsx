import { PluginStatusBadge } from "@/components/plugins/PluginStatusBadge";
import { WidgetEmpty, WidgetError, WidgetLoading, WidgetShell } from "@/components/dashboard/WidgetShell";
import { usePlugins } from "@/hooks/usePlugins";
import type { PluginStatus } from "@/lib/api";
import type { WidgetProps } from "@/lib/dashboard/types";

export function PluginsStatusWidget(_props: WidgetProps) {
  const { data, isLoading, isError } = usePlugins();

  if (isLoading) {
    return (
      <WidgetShell title="Plugins" deepLink="/plugins">
        <WidgetLoading />
      </WidgetShell>
    );
  }

  if (isError || !data) {
    return (
      <WidgetShell title="Plugins" deepLink="/plugins">
        <WidgetError />
      </WidgetShell>
    );
  }

  const plugins = data.plugins;
  if (plugins.length === 0) {
    return (
      <WidgetShell title="Plugins" deepLink="/plugins">
        <WidgetEmpty message="No plugins discovered." />
      </WidgetShell>
    );
  }

  const counts: Record<PluginStatus, number> = {
    running: 0,
    stopped: 0,
    error: 0,
  };
  for (const plugin of plugins) {
    counts[plugin.status] += 1;
  }
  const errorPlugins = plugins.filter((plugin) => plugin.status === "error").slice(0, 3);

  return (
    <WidgetShell title="Plugins" deepLink="/plugins">
      <div className="space-y-3">
        <div className="flex flex-wrap gap-3 text-sm">
          {(Object.keys(counts) as PluginStatus[]).map((status) => (
            <div key={status} className="flex items-center gap-2">
              <PluginStatusBadge status={status} />
              <span className="font-semibold text-ink">{counts[status]}</span>
            </div>
          ))}
        </div>
        {errorPlugins.length > 0 ? (
          <ul className="space-y-1 text-sm text-accent-rose">
            {errorPlugins.map((plugin) => (
              <li key={plugin.id} className="truncate">
                {plugin.name}
              </li>
            ))}
          </ul>
        ) : null}
      </div>
    </WidgetShell>
  );
}
