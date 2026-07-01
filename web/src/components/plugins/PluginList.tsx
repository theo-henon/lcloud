import type { PluginRecord } from "@/lib/api";
import { PluginStatusBadge } from "@/components/plugins/PluginStatusBadge";
import { Card } from "@/components/ui/card";
import { useTogglePlugin } from "@/hooks/usePlugins";

type PluginListProps = {
  plugins: PluginRecord[];
  isAdmin: boolean;
};

function EnableToggle({
  enabled,
  disabled,
  onChange,
}: {
  enabled: boolean;
  disabled?: boolean;
  onChange: (value: boolean) => void;
}) {
  return (
    <button
      type="button"
      role="switch"
      aria-checked={enabled}
      disabled={disabled}
      onClick={() => onChange(!enabled)}
      className={`relative h-6 w-11 rounded-full transition-colors disabled:cursor-not-allowed disabled:opacity-50 ${
        enabled ? "bg-primary" : "bg-surface-elevated"
      }`}
    >
      <span
        className={`absolute top-0.5 left-0.5 h-5 w-5 rounded-full transition-transform ${
          enabled ? "translate-x-5 bg-on-primary" : "bg-ink"
        }`}
      />
    </button>
  );
}

export function PluginList({ plugins, isAdmin }: PluginListProps) {
  const togglePlugin = useTogglePlugin();

  if (plugins.length === 0) {
    return (
      <Card className="p-6 text-sm text-muted">
        No plugins found — drop a binary and manifest into the plugins directory and restart
        lcloud.
      </Card>
    );
  }

  return (
    <div className="space-y-3">
      {plugins.map((plugin) => (
        <Card key={plugin.id} className="p-4">
          <div className="flex flex-wrap items-start justify-between gap-4">
            <div className="space-y-2">
              <div className="flex flex-wrap items-center gap-2">
                <h3 className="text-base font-semibold text-ink">{plugin.name}</h3>
                <PluginStatusBadge status={plugin.status} />
                <span className="font-mono text-xs text-muted">v{plugin.version}</span>
              </div>
              {plugin.description ? (
                <p className="text-sm text-body">{plugin.description}</p>
              ) : null}
              <p className="font-mono text-xs text-muted-soft">
                subscribes: {plugin.subscriptions.join(", ") || "—"}
              </p>
              {plugin.last_error ? (
                <p className="text-xs text-accent-rose">{plugin.last_error}</p>
              ) : null}
            </div>

            {isAdmin ? (
              <label className="flex items-center gap-2 text-sm text-body">
                <EnableToggle
                  enabled={plugin.enabled}
                  disabled={togglePlugin.isPending}
                  onChange={(enabled) => togglePlugin.mutate({ id: plugin.id, enabled })}
                />
                {plugin.enabled ? "Enabled" : "Disabled"}
              </label>
            ) : null}
          </div>
        </Card>
      ))}
    </div>
  );
}
